// linktest checks the phone server and USB link without touching routes.
// It uses its own phone path and port, so a running tunnel is not disturbed.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/hashicorp/yamux"
	M "github.com/sagernet/sing/common/metadata"
	"github.com/sagernet/sing/common/uot"
	"golang.org/x/net/proxy"

	"teltun/internal/wire"
)

const path = "/data/local/tmp/teltun-linktest"

var adbExe, serial string

func adb(args ...string) string {
	out, _ := exec.Command(adbExe, append([]string{"-s", serial}, args...)...).CombinedOutput()
	return strings.TrimSpace(string(out))
}

func main() {
	flag.StringVar(&adbExe, "adb", "adb", "")
	flag.StringVar(&serial, "s", "", "")
	bin := flag.String("bin", "", "phone binary")
	flag.Parse()

	fmt.Println("push:", adb("push", *bin, path))
	adb("shell", "chmod 755 "+path)
	ph := exec.Command(adbExe, "-s", serial, "shell", "exec "+path+" -listen 127.0.0.1:1082")
	ph.Stdout, ph.Stderr = os.Stdout, os.Stderr
	ph.Start()
	defer func() { ph.Process.Kill(); adb("shell", "pkill -f "+path); adb("forward", "--remove", "tcp:18082") }()
	time.Sleep(2 * time.Second)
	fmt.Println("forward:", adb("forward", "tcp:18082", "tcp:1082"))

	c, err := net.Dial("tcp", "127.0.0.1:18082")
	must(err)
	sess, err := yamux.Client(c, wire.Config())
	must(err)
	rtt, err := sess.Ping()
	must(err)
	fmt.Println("ping:", rtt)

	ln, _ := net.Listen("tcp", "127.0.0.1:0")
	go func() {
		for {
			lc, err := ln.Accept()
			if err != nil {
				return
			}
			go func() {
				s, err := sess.Open()
				if err != nil {
					lc.Close()
					return
				}
				go func() { io.Copy(s, lc); s.Close() }()
				io.Copy(lc, s)
				lc.Close()
			}()
		}
	}()
	d, _ := proxy.SOCKS5("tcp", ln.Addr().String(), nil, proxy.Direct)
	hc := &http.Client{Timeout: 60 * time.Second, Transport: &http.Transport{DialContext: d.(proxy.ContextDialer).DialContext}}

	r, err := hc.Get("https://api.ipify.org")
	must(err)
	ip, _ := io.ReadAll(r.Body)
	r.Body.Close()
	fmt.Println("public ip via phone:", string(ip))

	// download speed, 4 parallel streams inside the one link
	t0 := time.Now()
	done := make(chan int64)
	for i := 0; i < 4; i++ {
		go func() {
			r, err := hc.Get("https://speed.cloudflare.com/__down?bytes=25000000")
			if err != nil {
				done <- 0
				return
			}
			n, _ := io.Copy(io.Discard, r.Body)
			r.Body.Close()
			done <- n
		}()
	}
	var total int64
	for i := 0; i < 4; i++ {
		total += <-done
	}
	sec := time.Since(t0).Seconds()
	fmt.Printf("download: %.1f MB in %.1fs = %.0f Mbps\n", float64(total)/1e6, sec, float64(total)*8/1e6/sec)

	// UDP over TCP: DNS query to 1.1.1.1
	uc, err := d.Dial("tcp", uot.MagicAddress+":443")
	must(err)
	pc := uot.NewConn(uc, uot.Request{IsConnect: false})
	must(uot.WriteRequest(uc, uot.Request{IsConnect: false, Destination: M.ParseSocksaddr("1.1.1.1:53")}))
	q := []byte{0x12, 0x34, 1, 0, 0, 1, 0, 0, 0, 0, 0, 0, 7, 'e', 'x', 'a', 'm', 'p', 'l', 'e', 3, 'c', 'o', 'm', 0, 0, 1, 0, 1}
	_, err = pc.WriteTo(q, &net.UDPAddr{IP: net.IPv4(1, 1, 1, 1), Port: 53})
	must(err)
	uc.SetReadDeadline(time.Now().Add(5 * time.Second))
	buf := make([]byte, 1500)
	n, from, err := pc.ReadFrom(buf)
	must(err)
	fmt.Printf("udp dns reply: %d bytes from %v, id ok=%v\n", n, from, bytes.Equal(buf[:2], q[:2]))
}

func must(err error) {
	if err != nil {
		fmt.Println("FAIL:", err)
		os.Exit(1)
	}
}
