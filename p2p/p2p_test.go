package p2p

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"testing"

	"github.com/advancevillage/3rd/ecies"
)

var p2pTestData = map[string]struct {
	host     string
	tcpPort  int
	udpPort  int
	enodeUrl string
	pub      *ecdsa.PublicKey
	pri      *ecdsa.PrivateKey
	curve    elliptic.Curve
}{
	"case1": {
		curve:   elliptic.P256(),
		host:    "127.0.0.1",
		tcpPort: 13147,
		udpPort: 13147,
	},
}

func Test_p2p(t *testing.T) {
	for n, p := range p2pTestData {
		f := func(t *testing.T) {
			pri, err := ecdsa.GenerateKey(p.curve, rand.Reader)
			if err != nil {
				t.Fatal(err)
				return
			}
			//2. 生成enode
			var enode ecies.IENode
			enode, err = ecies.NewENodeByPP(pri, p.host, p.tcpPort, p.udpPort)
			if err != nil {
				t.Fatal(err)
				return
			}
			var ps IP2P
			ps, err = NewP2P(enode, nil)
			if err != nil {
				t.Fatal(err)
				return
			}
			ps.Start()
		}
		t.Run(n, f)
	}
}
