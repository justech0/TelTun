// teltun-phone runs on the Android device, started over `adb shell`.
//
// It listens on loopback only. The PC reaches it through `adb forward`,
// opens one multiplexed link, and sends SOCKS5 sessions inside it. Every
// outgoing connection is made by the phone, so to the network the traffic
// looks like ordinary phone traffic.
package main

import (
	"context"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/hashicorp/yamux"
	"github.com/sagernet/sing/common/buf"
	M "github.com/sagernet/sing/common/metadata"
	"github.com/sagernet/sing/common/uot"

	"teltun/internal/wire"
)

var dnsServer = flag.String("dns", "1.1.1.1:53", "DNS server the phone uses for hostnames")

func main() {
	listen := flag.String("listen", fmt.Sprintf("127.0.0.1:%d", wire.PhonePort), "listen address")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()
	if *showVersion {
		fmt.Println(wire.Version)
		return
	}
	log.SetFlags(log.Ltime)
	log.SetOutput(os.Stdout)
	setupResolver()

	ln, err := net.Listen("tcp", *listen)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	log.Printf("ready %s (v%s)", *listen, wire.Version)
	for {
		c, err := ln.Accept()
		if err != nil {
			log.Fatalf("accept: %v", err)
		}
		go serveLink(c)
	}
}

// Android has no /etc/resolv.conf, so Go's resolver needs a fixed server.
func setupResolver() {
	net.DefaultResolver = &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, _ string) (net.Conn, error) {
			var d net.Dialer
			return d.DialContext(ctx, network, *dnsServer)
		},
	}
}

func serveLink(c net.Conn) {
	if tc, ok := c.(*net.TCPConn); ok {
		tc.SetNoDelay(true)
	}
	sess, err := yamux.Server(c, wire.Config())
	if err != nil {
		log.Printf("link: %v", err)
		c.Close()
		return
	}
	log.Printf("pc connected")
	defer log.Printf("pc disconnected")
	for {
		s, err := sess.Accept()
		if err != nil {
			return
		}
		go handle(s)
	}
}

func handle(c net.Conn) {
	defer c.Close()
	c.SetDeadline(time.Now().Add(15 * time.Second))
	host, port, err := handshake(c)
	if err != nil {
		return
	}

	if host == uot.MagicAddress {
		reply(c, 0)
		c.SetDeadline(time.Time{})
		serveUDP(c)
		return
	}

	target := net.JoinHostPort(host, strconv.Itoa(port))
	out, err := net.DialTimeout("tcp", target, 10*time.Second)
	if err != nil {
		reply(c, 5) // connection refused
		return
	}
	defer out.Close()
	reply(c, 0)
	c.SetDeadline(time.Time{})
	pipe(c, out)
}

// handshake reads a SOCKS5 greeting and CONNECT request (no auth).
func handshake(c net.Conn) (string, int, error) {
	var h [2]byte
	if _, err := io.ReadFull(c, h[:]); err != nil || h[0] != 5 {
		return "", 0, errors.New("bad greeting")
	}
	if _, err := io.CopyN(io.Discard, c, int64(h[1])); err != nil {
		return "", 0, err
	}
	if _, err := c.Write([]byte{5, 0}); err != nil {
		return "", 0, err
	}

	var r [4]byte
	if _, err := io.ReadFull(c, r[:]); err != nil {
		return "", 0, err
	}
	if r[1] != 1 { // only CONNECT
		reply(c, 7)
		return "", 0, errors.New("unsupported command")
	}
	var host string
	switch r[3] {
	case 1:
		var a [4]byte
		if _, err := io.ReadFull(c, a[:]); err != nil {
			return "", 0, err
		}
		host = net.IP(a[:]).String()
	case 4:
		var a [16]byte
		if _, err := io.ReadFull(c, a[:]); err != nil {
			return "", 0, err
		}
		host = net.IP(a[:]).String()
	case 3:
		var n [1]byte
		if _, err := io.ReadFull(c, n[:]); err != nil {
			return "", 0, err
		}
		b := make([]byte, n[0])
		if _, err := io.ReadFull(c, b); err != nil {
			return "", 0, err
		}
		host = string(b)
	default:
		reply(c, 8)
		return "", 0, errors.New("bad address type")
	}
	var p [2]byte
	if _, err := io.ReadFull(c, p[:]); err != nil {
		return "", 0, err
	}
	return host, int(binary.BigEndian.Uint16(p[:])), nil
}

func reply(c net.Conn, code byte) {
	c.Write([]byte{5, code, 0, 1, 0, 0, 0, 0, 0, 0})
}

func pipe(a, b net.Conn) {
	var wg sync.WaitGroup
	wg.Add(2)
	cp := func(dst, src net.Conn) {
		defer wg.Done()
		io.Copy(dst, src)
		if cw, ok := dst.(interface{ CloseWrite() error }); ok {
			cw.CloseWrite()
		} else {
			dst.Close()
		}
	}
	go cp(a, b)
	go cp(b, a)
	wg.Wait()
}

// serveUDP handles sing-box "UDP over TCP" v2: UDP packets framed inside
// the stream. The phone sends them out from a real UDP socket.
func serveUDP(c net.Conn) {
	req, err := uot.ReadRequest(c)
	if err != nil {
		return
	}
	uc := uot.NewConn(c, *req)
	pc, err := net.ListenUDP("udp", nil)
	if err != nil {
		return
	}
	defer pc.Close()

	const idle = 2 * time.Minute
	done := make(chan struct{})
	go func() {
		defer close(done)
		b := make([]byte, 65535)
		for {
			pc.SetReadDeadline(time.Now().Add(idle))
			n, from, err := pc.ReadFromUDP(b)
			if err != nil {
				c.Close()
				return
			}
			if _, err := uc.WriteTo(b[:n], from); err != nil {
				return
			}
		}
	}()

	resolved := map[string]*net.UDPAddr{}
	for {
		b := buf.NewPacket()
		dst, err := uc.ReadPacket(b)
		if err != nil {
			b.Release()
			break
		}
		addr := udpAddr(dst, resolved)
		if addr != nil {
			pc.WriteToUDP(b.Bytes(), addr)
		}
		b.Release()
	}
	pc.Close()
	<-done
}

func udpAddr(d M.Socksaddr, cache map[string]*net.UDPAddr) *net.UDPAddr {
	if d.IsIP() {
		return d.UDPAddr()
	}
	key := d.String()
	if a, ok := cache[key]; ok {
		return a
	}
	a, err := net.ResolveUDPAddr("udp", key)
	if err != nil {
		return nil
	}
	cache[key] = a
	return a
}
