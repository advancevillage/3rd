//author: richard
package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/rsa"
	"errors"
	"fmt"
	"strings"
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
)

type ITokenClient interface {
	CreateToken(exp int) (string, error)
	ParseToken(token string) (bool, error)
}

type jwtHSClient struct {
	cli *jwt.Token
	exp int
	sct []byte
}

type jwtRSClient struct {
	cli        *jwt.Token
	exp        int
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

type jwtESClient struct {
	cli        *jwt.Token
	exp        int
	privateKey *ecdsa.PrivateKey
	publicKey  *ecdsa.PublicKey
}

func NewJwtClient(secret string, sm SignMethod) (ITokenClient, error) {
	var jsm jwt.SigningMethod
	switch sm {
	case RS256:
		jsm = jwt.SigningMethodRS256
		return newJwtRSClient(secret, jsm)
	case RS384:
		jsm = jwt.SigningMethodRS384
		return newJwtRSClient(secret, jsm)
	case RS512:
		jsm = jwt.SigningMethodRS512
		return newJwtRSClient(secret, jsm)
	case HS256:
		jsm = jwt.SigningMethodHS256
		return newJwtHSClient(secret, jsm)
	case HS384:
		jsm = jwt.SigningMethodHS384
		return newJwtHSClient(secret, jsm)
	case HS512:
		jsm = jwt.SigningMethodHS512
		return newJwtHSClient(secret, jsm)
	case ES256:
		jsm = jwt.SigningMethodES256
		return newJwtESClient(secret, jsm)
	case ES384:
		jsm = jwt.SigningMethodES384
		return newJwtESClient(secret, jsm)
	case ES512:
		jsm = jwt.SigningMethodES512
		return newJwtESClient(secret, jsm)
	default:
		return nil, fmt.Errorf("don't support %s method", sm)
	}
}

func newJwtHSClient(sct string, m jwt.SigningMethod) (*jwtHSClient, error) {
	var c = &jwtHSClient{}
	c.exp = 3600
	c.sct = []byte(sct)
	c.cli = jwt.New(m)
	return c, nil
}

//ssh-keygen -t rsa -P "" -b 4096 -m PEM -f jwtRS256.key
//openssl rsa -in jwtRS256.key -pubout -outform PEM -out jwtRS256.key.pub
func newJwtRSClient(sct string, m jwt.SigningMethod) (*jwtRSClient, error) {
	var c = &jwtRSClient{}
	var err error
	c.exp = 3600
	c.cli = jwt.New(m)
	//1. 分解公私密钥-
	var sp = strings.Split(sct, "|")
	if len(sp) != 2 {
		return nil, errors.New("don't have private key and public key")
	}
	var prk = strings.Trim(sp[0], "\n")
	prk = strings.Trim(prk, " ")
	var puk = strings.Trim(sp[1], "\n")
	puk = strings.Trim(puk, " ")
	//2. 验证有效性
	c.privateKey, err = jwt.ParseRSAPrivateKeyFromPEM([]byte(prk))
	if err != nil {
		return nil, err
	}
	c.publicKey, err = jwt.ParseRSAPublicKeyFromPEM([]byte(puk))
	if err != nil {
		return nil, err
	}
	return c, nil
}

func newJwtESClient(sct string, m jwt.SigningMethod) (*jwtESClient, error) {
	var c = &jwtESClient{}
	var err error
	c.exp = 3600
	c.cli = jwt.New(m)
	//1. 分解公私密钥-
	var sp = strings.Split(sct, "|")
	if len(sp) != 2 {
		return nil, errors.New("don't have private key and public key")
	}
	var prk = strings.Trim(sp[0], "\n")
	prk = strings.Trim(prk, " ")
	var puk = strings.Trim(sp[1], "\n")
	puk = strings.Trim(puk, " ")
	//2. 验证有效性
	c.privateKey, err = jwt.ParseECPrivateKeyFromPEM([]byte(prk))
	if err != nil {
		return nil, err
	}
	c.publicKey, err = jwt.ParseECPublicKeyFromPEM([]byte(puk))
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (c *jwtHSClient) CreateToken(exp int) (string, error) {
	defer func() {
		c.cli.Claims = jwt.MapClaims{}
	}()

	c.cli.Claims = &jwt.StandardClaims{
		Issuer:    "richard.sun",
		NotBefore: time.Now().Unix(),
		ExpiresAt: time.Now().Add(time.Duration(exp) * time.Second).Unix(),
	}

	return c.cli.SignedString(c.sct)
}

func (c *jwtHSClient) ParseToken(token string) (bool, error) {
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

func (c *jwtRSClient) CreateToken(exp int) (string, error) {
	defer func() {
		c.cli.Claims = jwt.MapClaims{}
	}()
	c.cli.Claims = &jwt.StandardClaims{
		Issuer:    "richard.sun",
		NotBefore: time.Now().Unix(),
		ExpiresAt: time.Now().Add(time.Duration(exp) * time.Second).Unix(),
	}
	return c.cli.SignedString(c.privateKey)
}

func (c *jwtRSClient) ParseToken(token string) (bool, error) {
	cb := func(tk *jwt.Token) (interface{}, error) {
		if _, ok := tk.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", tk.Header["alg"])
		}
		return c.publicKey, nil
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

func (c *jwtESClient) CreateToken(exp int) (string, error) {
	defer func() {
		c.cli.Claims = jwt.MapClaims{}
	}()

	c.cli.Claims = &jwt.StandardClaims{
		Issuer:    "richard.sun",
		NotBefore: time.Now().Unix(),
		ExpiresAt: time.Now().Add(time.Duration(exp) * time.Second).Unix(),
	}

	return c.cli.SignedString(c.privateKey)
}

func (c *jwtESClient) ParseToken(token string) (bool, error) {
	cb := func(tk *jwt.Token) (interface{}, error) {
		if _, ok := tk.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", tk.Header["alg"])
		}
		return c.publicKey, nil
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
