package engine

import (
	"context"
	"fmt"
	"path/filepath"

	box "github.com/sagernet/sing-box"
	"github.com/sagernet/sing-box/adapter/endpoint"
	"github.com/sagernet/sing-box/adapter/inbound"
	"github.com/sagernet/sing-box/adapter/outbound"
	"github.com/sagernet/sing-box/adapter/service"
	"github.com/sagernet/sing-box/dns"
	"github.com/sagernet/sing-box/dns/transport"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing-box/protocol/block"
	"github.com/sagernet/sing-box/protocol/direct"
	"github.com/sagernet/sing-box/protocol/socks"
	"github.com/sagernet/sing-box/protocol/tun"
	"github.com/sagernet/sing/common/json"
)

// boxContext registers only what TelTun uses: a TUN inbound and a SOCKS
// outbound. This keeps the exe much smaller than a full sing-box.
func boxContext() context.Context {
	in := inbound.NewRegistry()
	tun.RegisterInbound(in)

	out := outbound.NewRegistry()
	socks.RegisterOutbound(out)
	direct.RegisterOutbound(out)
	block.RegisterOutbound(out)

	d := dns.NewTransportRegistry()
	transport.RegisterTCP(d)
	transport.RegisterUDP(d)

	return box.Context(context.Background(), in, out, endpoint.NewRegistry(), d, service.NewRegistry())
}

// boxConfig builds the sing-box config: a virtual network card ("TelTun")
// takes all PC traffic and sends it to the local SOCKS port, which feeds
// the USB link. DNS also goes through the phone.
func boxConfig(socksPort int, blockQUIC bool) string {
	quic := ""
	if blockQUIC {
		// QUIC (UDP 443) over TCP is slow; browsers fall back to TCP.
		quic = `,{"network":"udp","port":443,"action":"reject"}`
	}
	logPath, _ := json.Marshal(filepath.Join(DataDir(), "sing-box.log"))
	return fmt.Sprintf(`{
  "log": {"level": "warn", "timestamp": true, "output": %s},
  "dns": {
    "servers": [
      {"type": "tcp", "tag": "dns-phone", "server": "1.1.1.1", "detour": "phone"},
      {"type": "tcp", "tag": "dns-phone-2", "server": "8.8.8.8", "detour": "phone"}
    ],
    "final": "dns-phone",
    "strategy": "ipv4_only"
  },
  "inbounds": [{
    "type": "tun", "tag": "tun-in", "interface_name": "TelTun",
    "address": ["172.19.0.1/30"], "mtu": 9000,
    "auto_route": true, "strict_route": true, "stack": "system"
  }],
  "outbounds": [
    {"type": "socks", "tag": "phone", "server": "127.0.0.1", "server_port": %d,
     "version": "5", "udp_over_tcp": {"enabled": true, "version": 2}},
    {"type": "direct", "tag": "direct"}
  ],
  "route": {
    "rules": [
      {"action": "sniff"},
      {"protocol": "dns", "action": "hijack-dns"}%s
    ],
    "final": "phone",
    "auto_detect_interface": false,
    "default_domain_resolver": "dns-phone"
  }
}`, logPath, socksPort, quic)
}

func startBox(socksPort int, blockQUIC bool) (*box.Box, error) {
	ctx := boxContext()
	opts, err := json.UnmarshalExtendedContext[option.Options](ctx, []byte(boxConfig(socksPort, blockQUIC)))
	if err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}
	b, err := box.New(box.Options{Context: ctx, Options: opts})
	if err != nil {
		return nil, err
	}
	if err := b.Start(); err != nil {
		b.Close()
		return nil, err
	}
	return b, nil
}
