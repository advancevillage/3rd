package netx

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/advancevillage/3rd/utils"
	"github.com/stretchr/testify/assert"
)

var macTestData = map[string]struct {
	fId    []byte
	flags  []byte
	expect []byte
	err    error
	hSize  int
	pad    int
	dLen   int
}{
	"case1": {
		hSize: 32,
		pad:   16,
		dLen:  16,
		flags: []byte{0x01, 0x0, 0x0, 0x0},
		fId:   utils.UUID8Byte(),
		err:   nil,
	},
	"case2": {
		hSize: 32,
		pad:   16,
		dLen:  10,
		flags: []byte{0x0, 0x0, 0x0, 0x0},
		fId:   utils.UUID8Byte(),
		err:   nil,
	},
	"case3": {
		hSize: 32,
		pad:   16,
		dLen:  21,
		flags: []byte{0x0, 0x0, 0x0, 0x0},
		fId:   utils.UUID8Byte(),
		err:   nil,
	},
	"case4": {
		hSize: 32,
		pad:   16,
		dLen:  32,
		flags: []byte{0x0, 0x0, 0x0, 0x0},
		fId:   utils.UUID8Byte(),
		err:   nil,
	},
	"case5": {
		hSize: 32,
		pad:   16,
		dLen:  63,
		flags: []byte{0x0, 0x0, 0x0, 0x0},
		fId:   utils.UUID8Byte(),
		err:   nil,
	},
	"case6": {
		hSize: 32,
		pad:   16,
		dLen:  127,
		flags: []byte{0x0, 0x0, 0x0, 0x0},
		fId:   utils.UUID8Byte(),
		err:   nil,
	},
	"case7": {
		hSize: 32,
		pad:   16,
		dLen:  512,
		flags: []byte{0x0, 0x0, 0x0, 0x0},
		fId:   utils.UUID8Byte(),
		err:   nil,
	},
	"case8": {
		hSize: 32,
		pad:   16,
		dLen:  1020,
		flags: []byte{0x0, 0x0, 0x0, 0x0},
		fId:   utils.UUID8Byte(),
		err:   nil,
	},
	"case9": {
		hSize: 32,
		pad:   16,
		dLen:  2021,
		flags: []byte{0x0, 0x0, 0x0, 0x0},
		fId:   utils.UUID8Byte(),
		err:   nil,
	},
	"case10": {
		hSize: 32,
		pad:   16,
		dLen:  4096,
		flags: []byte{0x0, 0x0, 0x0, 0x0},
		fId:   utils.UUID8Byte(),
		err:   nil,
	},
	"case11": {
		hSize: 32,
		pad:   16,
		dLen:  8011,
		flags: []byte{0x0, 0x0, 0x0, 0x0},
		fId:   utils.UUID8Byte(),
		err:   nil,
	},
	"case12": {
		hSize: 32,
		pad:   16,
		dLen:  16394,
		flags: []byte{0x0, 0x0, 0x0, 0x3},
		fId:   utils.UUID8Byte(),
		err:   nil,
	},
	"case13": {
		hSize: 32,
		pad:   16,
		dLen:  32788,
		flags: []byte{0x10, 0x0, 0x2, 0x0},
		fId:   utils.UUID8Byte(),
		err:   nil,
	},
	"case14": {
		hSize: 32,
		pad:   16,
		dLen:  65535,
		flags: []byte{0x01, 0x01, 0x0, 0x0},
		fId:   utils.UUID8Byte(),
		err:   nil,
	},
	"case15": {
		hSize: 32,
		pad:   16,
		dLen:  23434,
		flags: []byte{0x01, 0x01, 0x0, 0x0},
		fId:   utils.UUID8Byte(),
		err:   nil,
	},

	"case16": {
		hSize: 32,
		pad:   16,
		dLen:  113,
		flags: []byte{0x01, 0x01, 0x0, 0x0},
		fId:   utils.UUID8Byte(),
		err:   nil,
	},
	"case17": {
		hSize: 32,
		pad:   16,
		dLen:  3,
		flags: []byte{0x01, 0x01, 0x0, 0x0},
		fId:   utils.UUID8Byte(),
		err:   nil,
	},
	"case18": {
		hSize: 32,
		pad:   16,
		dLen:  281,
		flags: []byte{0x01, 0x01, 0x0, 0x0},
		fId:   utils.UUID8Byte(),
		err:   nil,
	},
	"case19": {
		hSize: 32,
		pad:   16,
		dLen:  1229,
		flags: []byte{0x01, 0x01, 0x0, 0x0},
		fId:   utils.UUID8Byte(),
		err:   nil,
	},
	"case20": {
		hSize: 32,
		pad:   16,
		dLen:  1723,
		flags: []byte{0x01, 0x01, 0x0, 0x0},
		fId:   utils.UUID8Byte(),
		err:   nil,
	},
}

func Test_frame(t *testing.T) {
	for n, p := range macTestData {
		key, _ := hex.DecodeString("6368616e676520746869732070617373776f726420746f206120736563726574")
		var sct = Secrets{
			MK:      key,
			AK:      key,
			Egress:  sha256.New(),
			Ingress: sha256.New(),
		}
		var m, err = NewTcpMac(sct)
		if err != nil {
			t.Fatal(err)
			return
		}
		f := func(t *testing.T) {
			//1. 设置期待值
			p.expect = []byte(utils.RandsString(p.dLen))
			var buf = make([]byte, 0, p.dLen<<1)
			var w = bytes.NewBuffer(buf)
			//2. 加密过程
			var err = m.WriteFrame(w, p.hSize, p.pad, p.flags, p.fId, p.expect)
			if err != nil {
				assert.Equal(t, p.err, err)
				return
			}
			//3. 解密过程
			var flags, fId, data []byte
			flags, fId, data, err = m.ReadFrame(w, p.hSize, p.pad)
			if err != nil {
				assert.Equal(t, p.err, err)
				return
			}
			//4. 验证过程
			assert.Equal(t, p.flags, flags)
			assert.Equal(t, p.fId, fId)
			assert.Equal(t, p.expect, data)
		}
		t.Run(n, f)
	}
}

//@overview: 单元测试 ecdhe 握手协议密钥交换

var ecdheTestData = map[string]struct {
	host    string
	tcpPort int
	udpPort int
	flags   []byte //加密帧  头信息
	fId     []byte
	dLen    int //数据长度
	expect  []byte
}{
	"case1": {
		host:    "127.0.0.1",
		tcpPort: 11011,
		udpPort: 11011,
		flags:   []byte{0x01, 0x01, 0x0, 0x0},
		fId:     utils.UUID8Byte(),
		dLen:    32,
	},
	"case2": {
		host:    "127.0.0.1",
		tcpPort: 11011,
		udpPort: 11011,
		flags:   []byte{0x01, 0x01, 0x0, 0x0},
		fId:     utils.UUID8Byte(),
		dLen:    1024,
	},
	"case3": {
		host:    "127.0.0.1",
		tcpPort: 11011,
		udpPort: 11011,
		flags:   []byte{0x01, 0x01, 0x0, 0x0},
		fId:     utils.UUID8Byte(),
		dLen:    4096,
	},
}

func Test_ecdhe(t *testing.T) {
	for n, p := range ecdheTestData {
		f := func(t *testing.T) {
			//0. 读写IO
			var rb = make([]byte, 0, 1024)
			var wb = make([]byte, 0, 1024)
			var r = bytes.NewBuffer(rb)
			var w = bytes.NewBuffer(wb)
			rPri, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
			if err != nil {
				t.Fatalf("generage receiver private key %s\n", err.Error())
				return
			}
			iPri, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
			if err != nil {
				t.Fatalf("generage initator private key %s\n", err.Error())
				return
			}
			//1. 构造接收端
			var receiver IECDHE
			receiver, err = NewECDHE256(rPri, p.host, p.tcpPort, p.udpPort)
			if err != nil {
				t.Fatalf("new receiver %s\n", err.Error())
				return
			}
			//2. 构造发送端
			var initator IECDHE
			initator, err = NewECDHE256(iPri, p.host, p.tcpPort, p.udpPort)
			if err != nil {
				t.Fatalf("new initator %s\n", err.Error())
				return
			}
			//3. 握手协议
			iRandPri, inonce, err := initator.Write(w, &rPri.PublicKey) //发送方发送握手协议报文
			if err != nil {
				t.Fatalf("initator send handshake package %s\n", err.Error())
				return
			}
			iPub, iRandPub, rInonce, err := receiver.Read(w) //接收方接收发送方 传输过来的 临时公钥和随机数
			if err != nil {
				t.Fatalf("receiver receive handshake package %s\n", err.Error())
				return
			}
			rRandPri, rnonce, err := receiver.Write(r, iPub) //接收方响应发送方
			if err != nil {
				t.Fatalf("receiver receive handshake package %s\n", err.Error())
				return
			}
			rPub, rRandPub, iRnonce, err := initator.Read(r) //发送方接收接收方的响应
			if err != nil {
				t.Fatalf("initator receive handshake package %s\n", err.Error())
				return
			}
			if !iPri.PublicKey.Equal(iPub) {
				t.Fatalf("iPri.PublicKey is not equal iPub")
				return
			}
			if !rPri.PublicKey.Equal(rPub) {
				t.Fatalf("rPri.PublicKey is not equal rPub")
				return
			}
			//5. 握手验证
			assert.Equal(t, inonce, rInonce)
			assert.Equal(t, rnonce, iRnonce)
			//6. 发送加解密消息
			imacSrt, err := initator.Ephemeral(iRandPri, inonce, rRandPub, iRnonce)
			if err != nil {
				t.Fatalf("initator init secret %s\n", err.Error())
				return
			}
			rmacSrt, err := receiver.Ephemeral(rRandPri, rnonce, iRandPub, rInonce)
			if err != nil {
				t.Fatalf("receiver init secret %s\n", err.Error())
				return
			}
			//7. 验证
			imac, err := NewTcpMac(*imacSrt)
			if err != nil {
				t.Fatalf("initator new tcp mac %s\n", err.Error())
				return
			}
			rmac, err := NewTcpMac(*rmacSrt)
			if err != nil {
				t.Fatalf("initator new tcp mac %s\n", err.Error())
				return
			}
			//8. 加密通讯
			p.expect = []byte(utils.RandsString(p.dLen))
			err = imac.WriteFrame(w, 0x20, 0x10, p.flags, p.fId, p.expect) //发送端发送消息
			if err != nil {
				assert.Equal(t, nil, err)
				return
			}
			err = imac.WriteFrame(w, 0x20, 0x10, p.flags, p.fId, p.expect) //发送端发送消息
			var flags, fId, data []byte
			flags, fId, data, err = rmac.ReadFrame(w, 0x20, 0x10) //接收端接收消息
			if err != nil {
				assert.Equal(t, nil, err)
				return
			}
			assert.Equal(t, flags, p.flags)
			assert.Equal(t, fId, p.fId)
			assert.Equal(t, data, p.expect)
		}
		t.Run(n, f)
	}
}
