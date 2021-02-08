package ecies

import (
	"crypto/elliptic"
	"crypto/rand"
	"testing"

	"github.com/advancevillage/3rd/utils"
	"github.com/stretchr/testify/assert"
)

var eciesTestData = map[string]struct {
	iPri *PriKey
	rPri *PriKey
	m    int
	salt []byte
}{
	"case1": {
		m:    12,
		salt: []byte(utils.RandsString(16)),
	},
	"case2": {
		m:    31,
		salt: []byte(utils.RandsString(15)),
	},
	"case3": {
		m:    64,
		salt: []byte(utils.RandsString(32)),
	},
	"case4": {
		m:    127,
		salt: []byte(utils.RandsString(16)),
	},
	"case5": {
		m:    255,
		salt: []byte(utils.RandsString(16)),
	},
	"case6": {
		m:    512,
		salt: []byte(utils.RandsString(16)),
	},
	"case7": {
		m:    12,
		salt: []byte(utils.RandsString(16)),
	},
	"case8": {
		m:    12,
		salt: nil,
	},
	"case9": {
		m:    512,
		salt: []byte(utils.RandsString(16)),
	},
	"case10": {
		m:    1023,
		salt: []byte(utils.RandsString(16)),
	},
	"case11": {
		m:    63,
		salt: nil,
	},
	"case12": {
		m:    127,
		salt: nil,
	},
}

func Test_ecies_p256(t *testing.T) {
	var ecies, err = NewECIES()
	if err != nil {
		t.Fatal(err)
		return
	}
	for n, p := range eciesTestData {
		f := func(t *testing.T) {
			//1. 生成发送方密钥对
			p.iPri, err = ecies.PriPub(rand.Reader, elliptic.P256())
			if err != nil {
				t.Fatal(err)
				return
			}
			//2. 生成接收方密钥对
			p.rPri, err = ecies.PriPub(rand.Reader, elliptic.P256())
			if err != nil {
				t.Fatal(err)
				return
			}
			//3. enode协议传输公钥信息
			var m = []byte(utils.RandsString(p.m))
			//4. 加密
			em, err := ecies.Encrypt(rand.Reader, &p.rPri.Pub, m, p.salt)
			if err != nil {
				t.Fatal(err)
				return
			}
			//5. 解密
			act, err := ecies.Decrypt(rand.Reader, p.rPri, em, p.salt)
			if err != nil {
				t.Fatal(err)
				return
			}
			//6. 验证
			assert.Equal(t, m, act)
		}
		t.Run(n, f)
	}
}

func Test_ecies_p384(t *testing.T) {
	var ecies, err = NewECIES()
	if err != nil {
		t.Fatal(err)
		return
	}
	for n, p := range eciesTestData {
		f := func(t *testing.T) {
			//1. 生成发送方密钥对
			p.iPri, err = ecies.PriPub(rand.Reader, elliptic.P384())
			if err != nil {
				t.Fatal(err)
				return
			}
			//2. 生成接收方密钥对
			p.rPri, err = ecies.PriPub(rand.Reader, elliptic.P384())
			if err != nil {
				t.Fatal(err)
				return
			}
			//3. enode协议传输公钥信息
			var m = []byte(utils.RandsString(p.m))
			//4. 加密
			em, err := ecies.Encrypt(rand.Reader, &p.rPri.Pub, m, p.salt)
			if err != nil {
				t.Fatal(err)
				return
			}
			//5. 解密
			act, err := ecies.Decrypt(rand.Reader, p.rPri, em, p.salt)
			if err != nil {
				t.Fatal(err)
				return
			}
			//6. 验证
			assert.Equal(t, m, act)
		}
		t.Run(n, f)
	}
}
