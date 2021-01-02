//author: richard
package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

//aes最少以16个字节分组，密钥key和随机数种子Iv和分组字节数相同是16个字节
const (
	sign = "kelly68chen.1995"
	iv   = "1995.chen86kelly"
)

//@link: https://juejin.im/post/5d2b0fbf51882547b2361a8a
func EncryptUseAes(plain []byte) ([]byte, error) {
	var block cipher.Block
	var err error
	if block, err = aes.NewCipher([]byte(sign)); err != nil {
		return nil, err
	}
	//创建ctr
	stream := cipher.NewCTR(block, []byte(iv))
	//加密, src,dst 可以为同一个内存地址
	stream.XORKeyStream(plain, plain)
	return plain, nil
}

func DecryptUseAes(cipher []byte) ([]byte, error) {
	//对密文再进行一次按位异或就可以得到明文
	//例如：3的二进制是0011和8的二进制1000按位异或(相同为0,不同为1)后得到1011，
	//对1011和8的二进制1000再进行按位异或得到0011即是3
	return EncryptUseAes(cipher)
}

//jwt token
type SignMethod string

const (
	RS256 = SignMethod("rs256")
	RS384 = SignMethod("rs384")
	RS512 = SignMethod("rs512")
	HS256 = SignMethod("hs256")
	HS384 = SignMethod("hs384")
	HS512 = SignMethod("hs512")
	ES256 = SignMethod("es256")
	ES384 = SignMethod("es384")
	ES512 = SignMethod("es512")
	PS256 = SignMethod("ps256")
	PS384 = SignMethod("ps384")
	PS512 = SignMethod("ps512")
)

type ITokenClient interface {
	CreateToken(exp int) (string, error)
	ParseToken(token string) (bool, error)
}

type jwtClient struct {
	cli *jwt.Token
	exp int
	sct []byte
}

func NewJwtClient(secret string, sm SignMethod) (ITokenClient, error) {
	var jsm jwt.SigningMethod
	switch sm {
	case RS256:
		jsm = jwt.SigningMethodRS256
	case RS384:
		jsm = jwt.SigningMethodRS384
	case RS512:
		jsm = jwt.SigningMethodRS512
	case HS256:
		jsm = jwt.SigningMethodHS256
	case HS384:
		jsm = jwt.SigningMethodHS384
	case HS512:
		jsm = jwt.SigningMethodHS512
	case ES256:
		jsm = jwt.SigningMethodES256
	case ES384:
		jsm = jwt.SigningMethodES384
	case ES512:
		jsm = jwt.SigningMethodES512
	case PS256:
		jsm = jwt.SigningMethodPS256
	case PS384:
		jsm = jwt.SigningMethodPS384
	case PS512:
		jsm = jwt.SigningMethodPS512
	default:
		return nil, fmt.Errorf("don't support %s method", sm)
	}

	var c = &jwtClient{}
	c.exp = 5
	c.sct = []byte(secret)
	c.cli = jwt.New(jsm)

	return c, nil
}

func (c *jwtClient) CreateToken(exp int) (string, error) {
	defer func() {
		c.cli.Claims = jwt.MapClaims{}
	}()

	c.cli.Claims = &jwt.StandardClaims{
		ExpiresAt: time.Now().Add(time.Duration(exp) * time.Second).Unix(),
	}

	return c.cli.SignedString(c.sct)
}

func (c *jwtClient) ParseToken(token string) (bool, error) {
	cb := func(tk *jwt.Token) (interface{}, error) {
		if _, ok := tk.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", tk.Header["alg"])
		}
		return c.sct, nil
	}

	var t, err = jwt.Parse(token, cb)
	if err != nil {
		return false, err
	}
	if t.Valid {
		return true, nil
	} else {
		return false, errors.New("invalid token string")
	}
}
