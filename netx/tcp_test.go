package netx

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"testing"
)

var tcpTestData = map[string]struct {
	host    string
	tcpPort int
	udpPort int
	pkgLen  int
	f       Handler
}{
	"case1": {
		host:    "192.168.187.176",
		tcpPort: 11211,
		udpPort: 11211,
		pkgLen:  32,
		f: func(ctx context.Context, in []byte) []byte {
			return in
		},
	},
}

func Test_tcp(t *testing.T) {
	for n, p := range tcpTestData {
		f := func(t *testing.T) {
			var pri, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
			if err != nil {
				t.Fatal(err)
				return
			}
			var so = &ServerOption{}
			so.Host = p.host
			so.Port = p.tcpPort
			so.UdpPort = p.udpPort
			so.PriKey = pri
			s, err := NewTCPServer(so, p.f)
			if err != nil {
				t.Fatal(err)
				return
			}
			s.StartServer()
		}
		t.Run(n, f)
	}
}
