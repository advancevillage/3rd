package netx

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/advancevillage/3rd/ecies"
	"github.com/advancevillage/3rd/utils"
	"github.com/stretchr/testify/assert"
)

var tcpTestData = map[string]struct {
	host    string
	tcpPort int
	udpPort int
	pkgLen  int
	f       Handler
}{
	"case1": {
		host:    "192.168.0.102",
		tcpPort: 1995,
		udpPort: 1995,
		pkgLen:  32,
		f: func(ctx context.Context, in []byte) []byte {
			return in
		},
	},
	"case2": {
		host:    "192.168.0.102",
		tcpPort: 1996,
		udpPort: 1996,
		pkgLen:  511,
		f: func(ctx context.Context, in []byte) []byte {
			return in
		},
	},
	"case3": {
		host:    "192.168.0.102",
		tcpPort: 1994,
		udpPort: 1994,
		pkgLen:  1023,
		f: func(ctx context.Context, in []byte) []byte {
			return in
		},
	},
}

func Test_tcp_server(t *testing.T) {
	for n, p := range tcpTestData {
		f := func(t *testing.T) {
			var pri, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
			if err != nil {
				t.Fatal(err)
				return
			}
			nodeCli, err := ecies.NewENodeByPP(pri, p.host, p.tcpPort, p.udpPort)
			if err != nil {
				t.Fatal(err)
				return
			}
			nodeUrl, err := nodeCli.GetENodeUrl()
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
			go s.StartServer()
			cPri, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
			if err != nil {
				t.Fatal(err)
				return
			}
			var co = &ClientOption{}
			co.EnodeUrl = nodeUrl
			co.PriKey = cPri
			c, err := NewTCPClient(co)
			if err != nil {
				t.Fatal(err)
				return
			}
			time.Sleep(5 * time.Second)
			var im = make(map[string]struct{})
			var rm = make(map[string]struct{})
			go func() {
				for len(im) < p.pkgLen {
					var msg = utils.RandsString(p.pkgLen)
					err := c.Send(context.TODO(), []byte(msg))
					if err != nil {
						return
					}
					im[msg] = struct{}{}
				}
			}()
			go func() {
				for len(rm) < p.pkgLen {
					buf, err := c.Receive(context.TODO())
					if err != nil {
						return
					}
					rm[string(buf)] = struct{}{}
				}
			}()
			//控制部分
			var ctlChan = make(chan struct{})
			var timeout = time.NewTicker(time.Minute)
			var sig = make(chan os.Signal)
			signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
			select {
			case <-ctlChan:
			case <-timeout.C:
			case <-sig:
			}
			for k := range im {
				delete(im, k)
				delete(rm, k)
			}
			assert.Equal(t, len(im), len(rm))
		}
		t.Run(n, f)
	}
}
