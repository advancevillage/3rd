package ecies

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

var enodeTestData = map[string]struct {
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
		host:    "192.168.1.101",
		tcpPort: 13147,
		udpPort: 13147,
	},
	"case2": {
		curve:   elliptic.P256(),
		host:    "192.168.1.101",
		tcpPort: 13147,
		udpPort: 13148,
	},
	"case3": {
		curve:   elliptic.P256(),
		host:    "192.168.1.101",
		tcpPort: 4321,
		udpPort: 4321,
	},
	"case4": {
		curve:   elliptic.P256(),
		host:    "192.168.1.101",
		tcpPort: 4321,
		udpPort: 4322,
	},
	"case5": {
		curve:   elliptic.P256(),
		host:    "127.0.0.1",
		tcpPort: 4321,
		udpPort: 4322,
	},
}

func Test_enode(t *testing.T) {
	for n, p := range enodeTestData {
		f := func(t *testing.T) {
			//1. 生成公私密钥对
			var err error
			p.pri, err = ecdsa.GenerateKey(p.curve, rand.Reader)
			if err != nil {
				t.Fatal(err)
				return
			}
			p.pub = &p.pri.PublicKey
			//2. 生成enode
			var enode IENode
			enode, err = NewENodeByPP(p.pri, p.host, p.tcpPort, p.udpPort)
			if err != nil {
				t.Fatal(err)
				return
			}
			p.enodeUrl, err = enode.GetENodeUrl()
			if err != nil {
				t.Fatal(err)
				return
			}
			//3. 逆过程
			var enode2 IENode
			enode2, err = NewENodeByUrl(p.enodeUrl)
			if err != nil {
				t.Fatal(err)
				return
			}
			//4. 验证
			assert.Equal(t, p.host, enode2.GetTcpHost())
			assert.Equal(t, p.tcpPort, enode2.GetTcpPort())
			assert.Equal(t, p.udpPort, enode2.GetUdpPort())
			assert.Equal(t, elliptic.Marshal(p.pub.Curve, p.pub.X, p.pub.Y), elliptic.Marshal(enode2.GetPubKey().Curve, enode2.GetPubKey().X, enode2.GetPubKey().Y))
		}
		t.Run(n, f)
	}
}
