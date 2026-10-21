// Package engine runs the tunnel: phone server over adb, the USB link,
// and the virtual network card (sing-box TUN).
package engine

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/yamux"
	box "github.com/sagernet/sing-box"
	"golang.org/x/net/proxy"

	"teltun/internal/wire"
)

const phonePath = "/data/local/tmp/teltun"

type State string

const (
	Idle          State = "idle"
	Connecting    State = "connecting"
	Connected     State = "connected"
	Disconnecting State = "disconnecting"
)

// Status is sent to the UI. Step and Error are i18n keys; Detail is raw text.
type Status struct {
	State    State  `json:"state"`
	Step     string `json:"step"`
	Error    string `json:"error"`
	Detail   string `json:"detail"`
	Device   string `json:"device"`
	PublicIP string `json:"publicIp"`
	Since    int64  `json:"since"`
}

type Stats struct {
	Down      float64 `json:"down"` // bytes per second
	Up        float64 `json:"up"`
	TotalDown uint64  `json:"totalDown"`
	TotalUp   uint64  `json:"totalUp"`
	Streams   int64   `json:"streams"`
}

type Device struct {
	Serial string `json:"serial"`
	State  string `json:"state"` // device, unauthorized, offline
	Model  string `json:"model"`
}

// StepError carries an i18n key for the UI plus the raw cause.
type StepError struct {
	Key string
	Err error
}

func (e *StepError) Error() string {
	if e.Err == nil {
		return e.Key
	}
	return e.Key + ": " + e.Err.Error()
}

func fail(key string, err error) error { return &StepError{Key: key, Err: err} }

type Engine struct {
	OnStatus func(Status)
	OnStats  func(Stats)
	OnLog    func(string)
	// OnLost fires when a live tunnel drops by itself (cable out, phone off).
	OnLost func()

	mu      sync.Mutex
	status  Status
	adbPath string
	runDir  string
	serial  string

	phone *exec.Cmd
	fwd   int // local adb forward port
	link  *link
	box   *box.Box
	stop  chan struct{}

	logMu   sync.Mutex
	logs    []string
	logFile *os.File
}

func New() *Engine {
	e := &Engine{status: Status{State: Idle}}
	os.MkdirAll(DataDir(), 0o755)
	e.logFile, _ = os.OpenFile(filepath.Join(DataDir(), "teltun.log"), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	return e
}

func (e *Engine) Logf(format string, a ...any) {
	line := time.Now().Format("15:04:05") + "  " + fmt.Sprintf(format, a...)
	e.logMu.Lock()
	e.logs = append(e.logs, line)
	if len(e.logs) > 500 {
		e.logs = e.logs[len(e.logs)-500:]
	}
	if e.logFile != nil {
		e.logFile.WriteString(line + "\r\n")
	}
	e.logMu.Unlock()
	if e.OnLog != nil {
		e.OnLog(line)
	}
}

func (e *Engine) Logs() []string {
	e.logMu.Lock()
	defer e.logMu.Unlock()
	return append([]string(nil), e.logs...)
}

func (e *Engine) Status() Status {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.status
}

func (e *Engine) setStatus(f func(*Status)) {
	e.mu.Lock()
	f(&e.status)
	s := e.status
	e.mu.Unlock()
	if e.OnStatus != nil {
		e.OnStatus(s)
	}
}

func (e *Engine) step(key string) {
	e.Logf("» %s", key)
	e.setStatus(func(s *Status) { s.Step = key })
}

func (e *Engine) prepare() error {
	if e.adbPath != "" {
		return nil
	}
	dir, err := unpack()
	if err != nil {
		return fail("errPrepare", err)
	}
	e.runDir = dir
	e.adbPath = filepath.Join(dir, "adb", "adb.exe")
	return nil
}

// Devices lists phones adb can see.
func (e *Engine) Devices() ([]Device, error) {
	if err := e.prepare(); err != nil {
		return nil, err
	}
	out, err := e.runAdb("", 20*time.Second, "devices", "-l")
	if err != nil {
		return nil, fail("errAdb", fmt.Errorf("%v: %s", err, out))
	}
	var list []Device
	for _, ln := range strings.Split(out, "\n") {
		f := strings.Fields(ln)
		if len(f) < 2 || strings.HasPrefix(ln, "List of") || strings.HasPrefix(ln, "*") {
			continue
		}
		d := Device{Serial: f[0], State: f[1]}
		for _, kv := range f[2:] {
			if m, ok := strings.CutPrefix(kv, "model:"); ok {
				d.Model = strings.ReplaceAll(m, "_", " ")
			}
		}
		list = append(list, d)
	}
	return list, nil
}

// Connect brings the tunnel up. serial may be empty to pick the only phone.
func (e *Engine) Connect(serial string, blockQUIC bool) error {
	e.mu.Lock()
	if e.status.State != Idle {
		e.mu.Unlock()
		return errors.New("busy")
	}
	e.status = Status{State: Connecting}
	e.mu.Unlock()
	e.setStatus(func(*Status) {})

	err := e.connect(serial, blockQUIC)
	if err != nil {
		e.Logf("error: %v", err)
		e.teardown()
		key, detail := "errUnknown", err.Error()
		var se *StepError
		if errors.As(err, &se) {
			key = se.Key
			if se.Err != nil {
				detail = se.Err.Error()
			} else {
				detail = ""
			}
		}
		e.setStatus(func(s *Status) { *s = Status{State: Idle, Error: key, Detail: detail} })
		return err
	}
	return nil
}

func (e *Engine) connect(serial string, blockQUIC bool) error {
	e.step("stepPrepare")
	if err := e.prepare(); err != nil {
		return err
	}

	e.step("stepDevice")
	devs, err := e.Devices()
	if err != nil {
		return err
	}
	var dev *Device
	for i := range devs {
		if serial == "" || devs[i].Serial == serial {
			dev = &devs[i]
			break
		}
	}
	switch {
	case dev == nil:
		return fail("errNoDevice", nil)
	case dev.State == "unauthorized":
		return fail("errUnauthorized", nil)
	case dev.State != "device":
		return fail("errOffline", errors.New(dev.State))
	}
	e.mu.Lock()
	e.serial = dev.Serial
	e.mu.Unlock()
	name := dev.Model
	if name == "" {
		name = dev.Serial
	}
	e.setStatus(func(s *Status) { s.Device = name })
	e.Logf("device: %s (%s)", name, dev.Serial)

	e.step("stepPush")
	arch, _ := e.adb(10*time.Second, "shell", "uname -m")
	bin := "teltun-arm64"
	switch {
	case strings.Contains(arch, "aarch64"), strings.Contains(arch, "arm64"):
	case strings.HasPrefix(arch, "arm"):
		bin = "teltun-arm"
	case arch == "x86_64": // emulators
		bin = "teltun-amd64"
	}
	e.Logf("phone cpu: %s -> %s", arch, bin)
	e.adb(10*time.Second, "shell", "pkill -f "+phonePath)
	if out, err := e.adb(60*time.Second, "push", filepath.Join(e.runDir, "phone", bin), phonePath); err != nil {
		return fail("errPush", fmt.Errorf("%v: %s", err, out))
	}
	if out, err := e.adb(10*time.Second, "shell", "chmod 755 "+phonePath); err != nil {
		return fail("errPush", fmt.Errorf("%v: %s", err, out))
	}

	e.step("stepPhone")
	if err := e.startPhone(); err != nil {
		return err
	}

	e.step("stepLink")
	out, err := e.adb(10*time.Second, "forward", "tcp:0", fmt.Sprintf("tcp:%d", wire.PhonePort))
	if err != nil {
		return fail("errForward", fmt.Errorf("%v: %s", err, out))
	}
	fwd, err := strconv.Atoi(strings.TrimSpace(out))
	if err != nil {
		return fail("errForward", fmt.Errorf("bad port %q", out))
	}
	e.mu.Lock()
	e.fwd = fwd
	e.mu.Unlock()
	e.Logf("adb forward 127.0.0.1:%d -> phone:%d", fwd, wire.PhonePort)
	if err := e.openLink(fwd); err != nil {
		return err
	}

	e.step("stepTest")
	ip, err := e.selfTest()
	if err != nil {
		return fail("errTest", err)
	}
	e.Logf("internet via phone OK")
	e.setStatus(func(s *Status) { s.PublicIP = ip })

	e.step("stepTun")
	b, err := startBox(e.link.ln.Addr().(*net.TCPAddr).Port, blockQUIC)
	if err != nil {
		return fail("errTun", err)
	}
	e.mu.Lock()
	e.box = b
	e.stop = make(chan struct{})
	e.mu.Unlock()

	go e.watch(e.link, e.stop)
	e.Logf("tunnel up")
	e.setStatus(func(s *Status) { s.State = Connected; s.Step = ""; s.Since = time.Now().Unix() })
	return nil
}

func (e *Engine) startPhone() error {
	args := []string{"-s", e.serial, "shell", "exec " + phonePath}
	cmd := hidden(exec.Command(e.adbPath, args...))
	pr, pw, _ := os.Pipe()
	cmd.Stdout = pw
	cmd.Stderr = pw
	if err := cmd.Start(); err != nil {
		pw.Close()
		return fail("errPhone", err)
	}
	pw.Close()
	e.mu.Lock()
	e.phone = cmd
	e.mu.Unlock()

	ready := make(chan struct{}, 1)
	go func() {
		sc := bufio.NewScanner(pr)
		for sc.Scan() {
			line := strings.TrimSpace(sc.Text())
			e.Logf("phone: %s", line)
			if strings.Contains(line, "ready") {
				select {
				case ready <- struct{}{}:
				default:
				}
			}
		}
	}()
	select {
	case <-ready:
		return nil
	case <-time.After(10 * time.Second):
		return fail("errPhone", errors.New("phone server did not start"))
	}
}

func (e *Engine) openLink(port int) error {
	var lastErr error
	for i := 0; i < 5; i++ {
		c, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 5*time.Second)
		if err != nil {
			lastErr = err
			time.Sleep(500 * time.Millisecond)
			continue
		}
		c.(*net.TCPConn).SetNoDelay(true)
		sess, err := yamux.Client(c, wire.Config())
		if err != nil {
			c.Close()
			lastErr = err
			continue
		}
		if _, err := sess.Ping(); err != nil {
			sess.Close()
			lastErr = err
			time.Sleep(500 * time.Millisecond)
			continue
		}
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			sess.Close()
			return fail("errLink", err)
		}
		l := &link{sess: sess, ln: ln}
		l.wg.Add(1)
		go l.serve()
		e.mu.Lock()
		e.link = l
		e.mu.Unlock()
		return nil
	}
	return fail("errLink", lastErr)
}

// selfTest checks the phone can reach the internet before we take over
// the PC's routes. It also learns the public IP for the UI.
func (e *Engine) selfTest() (string, error) {
	d, err := proxy.SOCKS5("tcp", e.link.ln.Addr().String(), nil, proxy.Direct)
	if err != nil {
		return "", err
	}
	cd := d.(proxy.ContextDialer)
	client := &http.Client{
		Timeout:   15 * time.Second,
		Transport: &http.Transport{DialContext: cd.DialContext},
	}
	var lastErr error
	for _, u := range []string{"http://connectivitycheck.gstatic.com/generate_204", "http://cp.cloudflare.com/generate_204"} {
		resp, err := client.Get(u)
		if err != nil {
			lastErr = err
			continue
		}
		resp.Body.Close()
		if resp.StatusCode == 204 || resp.StatusCode == 200 {
			lastErr = nil
			break
		}
		lastErr = fmt.Errorf("%s: http %d", u, resp.StatusCode)
	}
	if lastErr != nil {
		return "", lastErr
	}
	ip := ""
	if resp, err := client.Get("https://api.ipify.org"); err == nil {
		b := make([]byte, 64)
		n, _ := resp.Body.Read(b)
		resp.Body.Close()
		ip = strings.TrimSpace(string(b[:n]))
	}
	return ip, nil
}

// watch reports speed every second and notices when the link dies.
func (e *Engine) watch(l *link, stop chan struct{}) {
	t := time.NewTicker(time.Second)
	defer t.Stop()
	var lastD, lastU uint64
	last := time.Now()
	for {
		select {
		case <-stop:
			return
		case <-l.sess.CloseChan():
			e.Logf("link lost (cable or phone)")
			go func() {
				e.Disconnect()
				if e.OnLost != nil {
					e.OnLost()
				}
			}()
			return
		case now := <-t.C:
			d, u := l.down.Load(), l.up.Load()
			sec := now.Sub(last).Seconds()
			if e.OnStats != nil {
				e.OnStats(Stats{
					Down: float64(d-lastD) / sec, Up: float64(u-lastU) / sec,
					TotalDown: d, TotalUp: u, Streams: l.streams.Load(),
				})
			}
			lastD, lastU, last = d, u, now
		}
	}
}

// Disconnect tears everything down and gives the PC its normal network back.
func (e *Engine) Disconnect() {
	e.mu.Lock()
	if e.status.State != Connected {
		e.mu.Unlock()
		return
	}
	e.mu.Unlock()
	e.setStatus(func(s *Status) { s.State = Disconnecting })
	e.teardown()
	e.Logf("tunnel down")
	e.setStatus(func(s *Status) { *s = Status{State: Idle} })
}

func (e *Engine) teardown() {
	e.mu.Lock()
	b, l, ph, stop, fwd := e.box, e.link, e.phone, e.stop, e.fwd
	e.box, e.link, e.phone, e.stop, e.fwd = nil, nil, nil, nil, 0
	e.mu.Unlock()

	if stop != nil {
		close(stop)
	}
	if b != nil {
		b.Close() // removes the TelTun card and its routes
	}
	if l != nil {
		l.close()
	}
	if ph != nil && ph.Process != nil {
		ph.Process.Kill()
		ph.Wait()
	}
	if e.adbPath != "" && e.serial != "" {
		if fwd != 0 {
			e.adb(5*time.Second, "forward", "--remove", fmt.Sprintf("tcp:%d", fwd))
		}
		e.adb(5*time.Second, "shell", "pkill -f "+phonePath)
	}
}

// Shutdown is called when the app quits.
func (e *Engine) Shutdown() {
	e.Disconnect()
	if e.adbPath != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		hidden(exec.CommandContext(ctx, e.adbPath, "kill-server")).Run()
	}
	if e.logFile != nil {
		e.logFile.Close()
	}
}
