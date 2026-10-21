// Package wire holds what the PC and the phone must agree on.
//
// The PC talks to the phone over ONE TCP connection that adb forwards
// through USB. Many adb streams at once stall, so every connection the PC
// makes is a yamux stream inside that single link. Each stream carries a
// plain SOCKS5 session; the phone side dials the real target.
package wire

import (
	"io"
	"time"

	"github.com/hashicorp/yamux"
)

// Version is bumped when the PC <-> phone protocol changes.
const Version = "2"

// PhonePort is the loopback port the phone server listens on.
const PhonePort = 1081

// Config returns the yamux settings used on both ends.
func Config() *yamux.Config {
	c := yamux.DefaultConfig()
	c.AcceptBacklog = 1024
	c.MaxStreamWindowSize = 16 << 20 // big window keeps the single USB link full
	c.KeepAliveInterval = 10 * time.Second
	c.ConnectionWriteTimeout = 30 * time.Second
	c.StreamOpenTimeout = 15 * time.Second
	c.StreamCloseTimeout = 2 * time.Minute
	c.LogOutput = io.Discard
	return c
}
