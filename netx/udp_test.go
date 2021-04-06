package netx

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/advancevillage/3rd/utils"
	"github.com/stretchr/testify/assert"
)

var udpTestData = map[string]struct {
	host    string
	udpPort int
	pkgLen  int
	f       Handler
}{}

//go test -v -count=1 -timeout 20m -test.run Test_udp_server ./netx/
func Test_udp_server(t *testing.T) {
	var host = "127.0.0.1"
	var port = 1994
	var hf = func(ctx context.Context, in []byte) []byte {
		return in
	}
	var primes = []int{
		2, 3, 5, 7, 11, 13, 17, 19, 23, 29,
	}
	//初始化数据
	for k, v := range primes {
		udpTestData[fmt.Sprintf("case%d", k)] = struct {
			host    string
			udpPort int
			pkgLen  int
			f       Handler
		}{
			host:    host,
			udpPort: port,
			pkgLen:  v,
			f:       hf,
		}
	}

	var scfg = &ServerOption{}
	scfg.Host = host
	scfg.UdpPort = port
	scfg.MaxSize = 1 << 16
	var udps, err = NewUDPServer(scfg, hf)
	if err != nil {
		t.Fatal(err)
		return
	}
	//2. 启动服务
	go udps.StartServer()
	time.Sleep(5 * time.Second)

	for n, p := range udpTestData {
		f := func(t *testing.T) {
			var ccfg = &ClientOption{}
			ccfg.Host = p.host
			ccfg.UdpPort = p.udpPort
			ccfg.MaxSize = uint32(p.pkgLen + 64)
			var udpc IUDPClient
			udpc, err = NewUDPClient(ccfg)
			if err != nil {
				t.Fatal(err)
				return
			}
			var im = make(map[string]struct{})
			var rm = make(map[string]struct{})
			var iChan = make(chan struct{})
			var rChan = make(chan struct{})
			go func() {
				for len(im) < p.pkgLen {
					var msg = utils.RandsString(p.pkgLen)
					err := udpc.Send(context.TODO(), []byte(msg))
					if err != nil {
						continue
					}
					im[msg] = struct{}{}
				}
				iChan <- struct{}{}
			}()
			go func() {
				for len(rm) < p.pkgLen {
					buf, err := udpc.Receive(context.TODO())
					if err != nil {
						continue
					}
					rm[string(buf)] = struct{}{}
				}
				rChan <- struct{}{}
			}()
			//控制部分
			select {
			case <-iChan:
				<-rChan
			case <-rChan:
				<-iChan
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
