package main

import (
	"context"
	"os/exec"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"teltun/internal/autostart"
	"teltun/internal/engine"
	"teltun/internal/hotspot"
	"teltun/internal/settings"
)

// App is what the web UI can call (bound by Wails).
type App struct {
	ctx      context.Context
	eng      *engine.Engine
	tray     *tray
	quitting bool

	mu       sync.Mutex
	settings settings.Settings
	devices  []engine.Device
	manual   bool // user pressed Disconnect; do not auto-reconnect
	lost     bool // link dropped by itself; reconnect when phone is back
}

type Snapshot struct {
	Version   string            `json:"version"`
	Status    engine.Status     `json:"status"`
	Settings  settings.Settings `json:"settings"`
	Autostart bool              `json:"autostart"`
	Devices   []engine.Device   `json:"devices"`
	Logs      []string          `json:"logs"`
}

func NewApp() *App {
	return &App{eng: engine.New(), settings: settings.Load()}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.eng.OnStatus = func(s engine.Status) {
		runtime.EventsEmit(ctx, "status", s)
		a.tray.update(s.State, a.lang())
	}
	a.eng.OnStats = func(s engine.Stats) { runtime.EventsEmit(ctx, "stats", s) }
	a.eng.OnLog = func(l string) { runtime.EventsEmit(ctx, "log", l) }
	a.eng.OnLost = func() {
		a.mu.Lock()
		a.lost = true
		a.mu.Unlock()
		runtime.EventsEmit(ctx, "lost")
	}
	a.eng.Logf("TelTun %s started", version)
	go a.deviceLoop()
}

func (a *App) shutdown(ctx context.Context) {
	a.eng.Shutdown()
}

// beforeClose hides to tray instead of quitting, if the user wants that.
func (a *App) beforeClose(ctx context.Context) bool {
	if a.quitting || !a.Settings().CloseToTray {
		return false
	}
	runtime.WindowHide(ctx)
	return true
}

func (a *App) lang() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.settings.Language
}

// deviceLoop watches for phones and handles auto connect / reconnect.
func (a *App) deviceLoop() {
	first := true
	for {
		if a.eng.Status().State == engine.Idle {
			devs, err := a.eng.Devices()
			if err == nil {
				a.mu.Lock()
				a.devices = devs
				s := a.settings
				manual := a.manual
				a.mu.Unlock()
				runtime.EventsEmit(a.ctx, "devices", devs)

				ready := false
				for _, d := range devs {
					if d.State == "device" && (s.Device == "" || d.Serial == s.Device) {
						ready = true
					}
				}
				want := (first && s.AutoConnect) || (s.AutoReconnect && !manual && a.peekLost())
				if ready && want {
					a.wasLost()
					a.Connect()
				}
			}
			first = false
		}
		time.Sleep(3 * time.Second)
	}
}

func (a *App) peekLost() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.lost
}

func (a *App) wasLost() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	v := a.lost
	a.lost = false
	return v
}

// ---- bound methods ----

func (a *App) Snapshot() Snapshot {
	a.mu.Lock()
	defer a.mu.Unlock()
	return Snapshot{
		Version:   version,
		Status:    a.eng.Status(),
		Settings:  a.settings,
		Autostart: autostart.Enabled(),
		Devices:   a.devices,
		Logs:      a.eng.Logs(),
	}
}

func (a *App) Settings() settings.Settings {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.settings
}

func (a *App) SaveSettings(s settings.Settings) error {
	a.mu.Lock()
	a.settings = s
	a.mu.Unlock()
	a.tray.update(a.eng.Status().State, s.Language)
	return settings.Save(s)
}

func (a *App) SetAutostart(on bool) error {
	if on {
		return autostart.Enable()
	}
	return autostart.Disable()
}

func (a *App) Devices() ([]engine.Device, error) { return a.eng.Devices() }

// Connect starts the tunnel in the background; progress comes as events.
func (a *App) Connect() {
	a.mu.Lock()
	a.manual = false
	s := a.settings
	a.mu.Unlock()
	go func() {
		if err := a.eng.Connect(s.Device, s.BlockQUIC); err == nil && s.HotspotWithTun {
			a.HotspotStart()
		}
	}()
}

func (a *App) Disconnect() {
	a.mu.Lock()
	a.manual = true
	a.mu.Unlock()
	go a.eng.Disconnect()
}

func (a *App) Toggle() {
	if a.eng.Status().State == engine.Connected {
		a.Disconnect()
	} else if a.eng.Status().State == engine.Idle {
		a.Connect()
	}
}

func (a *App) HotspotStatus() (hotspot.Info, error) { return hotspot.Status() }

func (a *App) HotspotStart() (hotspot.Info, error) {
	s := a.Settings()
	i, err := hotspot.Start(s.HotspotSSID, s.HotspotPass)
	a.emitHotspot(i, err)
	return i, err
}

func (a *App) HotspotStop() (hotspot.Info, error) {
	i, err := hotspot.Stop()
	a.emitHotspot(i, err)
	return i, err
}

func (a *App) emitHotspot(i hotspot.Info, err error) {
	if err != nil {
		a.eng.Logf("hotspot: %v", err)
	} else {
		a.eng.Logf("hotspot: %s %s", i.State, i.SSID)
	}
	runtime.EventsEmit(a.ctx, "hotspot", i)
}

func (a *App) OpenLogFolder() {
	exec.Command("explorer.exe", engine.DataDir()).Start()
}

func (a *App) OpenURL(u string) { runtime.BrowserOpenURL(a.ctx, u) }

func (a *App) Show() {
	runtime.WindowShow(a.ctx)
	runtime.WindowUnminimise(a.ctx)
}

func (a *App) Hide()     { runtime.WindowHide(a.ctx) }
func (a *App) Minimise() { runtime.WindowMinimise(a.ctx) }

func (a *App) Quit() {
	a.quitting = true
	runtime.Quit(a.ctx)
}
