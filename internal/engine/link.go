package engine

import (
	"io"
	"net"
	"sync"
	"sync/atomic"

	"github.com/hashicorp/yamux"
)

// link is the PC end of the tunnel. It accepts local SOCKS connections
// (from sing-box) and sends each one as a stream over the single USB link.
type link struct {
	sess    *yamux.Session
	ln      net.Listener
	down    atomic.Uint64 // phone -> PC bytes
	up      atomic.Uint64 // PC -> phone bytes
	streams atomic.Int64
	wg      sync.WaitGroup
}

func (l *link) serve() {
	defer l.wg.Done()
	for {
		c, err := l.ln.Accept()
		if err != nil {
			return
		}
		go l.handle(c)
	}
}

func (l *link) handle(c net.Conn) {
	defer c.Close()
	s, err := l.sess.Open()
	if err != nil {
		return
	}
	defer s.Close()
	l.streams.Add(1)
	defer l.streams.Add(-1)

	done := make(chan struct{}, 2)
	go func() {
		io.Copy(&counter{w: s, n: &l.up}, c)
		s.Close()
		done <- struct{}{}
	}()
	go func() {
		io.Copy(&counter{w: c, n: &l.down}, s)
		if tc, ok := c.(*net.TCPConn); ok {
			tc.CloseWrite()
		}
		done <- struct{}{}
	}()
	<-done
	<-done
}

func (l *link) close() {
	l.ln.Close()
	l.sess.Close()
	l.wg.Wait()
}

type counter struct {
	w io.Writer
	n *atomic.Uint64
}

func (c *counter) Write(p []byte) (int, error) {
	n, err := c.w.Write(p)
	c.n.Add(uint64(n))
	return n, err
}
