package sm4

import (
	"crypto/cipher"
	"strconv"
)

type sm4 struct {
	enc []uint32
	dec []uint32
}

type KeySizeError int

func (k KeySizeError) Error() string {
	return "cryptox/sm4: invalid key size " + strconv.Itoa(int(k))
}

//GB/T 32907-2016; SM4-128
func NewCipher(key []byte) (cipher.Block, error) {
	k := len(key)
	switch k {
	default:
		return nil, KeySizeError(k)
	case 16:
		break
	}
	return newCipher(key)
}

func newCipher(key []byte) (cipher.Block, error) {
	c := sm4{}
	c.enc = expandEncKeyGo(key)
	c.dec = expandDecKeyGo(key)
	return &c, nil
}

func (c *sm4) BlockSize() int { return BlockSize }

func (c *sm4) Encrypt(dst, src []byte) {
	if len(src) < BlockSize {
		panic("crypto/sm4: input not full block")
	}
	if len(dst) < BlockSize {
		panic("crypto/sm4: output not full block")
	}
	if inexactOverlap(dst[:BlockSize], src[:BlockSize]) {
		panic("crypto/sm4: invalid buffer overlap")
	}
	encryptBlockGo(c.enc, dst, src)
}

func (c *sm4) Decrypt(dst, src []byte) {
	if len(src) < BlockSize {
		panic("crypto/sm4: input not full block")
	}
	if len(dst) < BlockSize {
		panic("crypto/sm4: output not full block")
	}
	if inexactOverlap(dst[:BlockSize], src[:BlockSize]) {
		panic("crypto/sm4: invalid buffer overlap")
	}
	encryptBlockGo(c.dec, dst, src)
}
