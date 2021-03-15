package netx

import (
	"context"
	"testing"
)

var udpTestData = map[string]struct {
	host    string
	udpPort int
	pkgLen  int
	f       Handler
}{
	"case1": {
		host:    "192.168.187.176",
		udpPort: 1994,
		pkgLen:  32,
		f: func(ctx context.Context, in []byte) []byte {
			return in
		},
	},
}

func Test_udp_server(t *testing.T) {
	for n, p := range udpTestData {
		f := func(t *testing.T) {
			var cfg = &ServerOption{}
			cfg.Host = p.host
			cfg.UdpPort = p.udpPort
			var udps, err = NewUDPServer(cfg, p.f)
			if err != nil {
				t.Fatal(err)
				return
			}
			udps.StartServer()
		}
		t.Run(n, f)
	}
}
