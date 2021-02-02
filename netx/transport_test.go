package netx

import (
	"bytes"
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
}

func Test_mac(t *testing.T) {
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
