package ecies

import (
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"hash"
	"io"
	"math/big"

	"github.com/advancevillage/3rd/utils"
	"golang.org/x/crypto/hkdf"
)

var (
	ECIES_AES128_SHA256 = &ECIESParams{
		Hash:      sha256.New,
		hashAlgo:  crypto.SHA256,
		Cipher:    aes.NewCipher,
		BlockSize: aes.BlockSize,
		KeyLen:    16,
	}

	ECIES_AES256_SHA256 = &ECIESParams{
		Hash:      sha256.New,
		hashAlgo:  crypto.SHA256,
		Cipher:    aes.NewCipher,
		BlockSize: aes.BlockSize,
		KeyLen:    32,
	}

	ECIES_AES256_SHA384 = &ECIESParams{
		Hash:      sha512.New384,
		hashAlgo:  crypto.SHA384,
		Cipher:    aes.NewCipher,
		BlockSize: aes.BlockSize,
		KeyLen:    32,
	}

	ECIES_AES256_SHA512 = &ECIESParams{
		Hash:      sha512.New,
		hashAlgo:  crypto.SHA512,
		Cipher:    aes.NewCipher,
		BlockSize: aes.BlockSize,
		KeyLen:    32,
	}
)

var (
	ErrImport                     = fmt.Errorf("ecies: failed to import key")
	ErrInvalidMessage             = fmt.Errorf("ecies: invalid message")
	ErrInvalidCurve               = fmt.Errorf("ecies: invalid elliptic curve")
	ErrInvalidPublicKey           = fmt.Errorf("ecies: invalid public key")
	ErrSharedKeyIsPointAtInfinity = fmt.Errorf("ecies: shared key is point at infinity")
	ErrSharedKeyTooBig            = fmt.Errorf("ecies: shared key params are too big")

	ErrUnsupportedECDHAlgorithm   = fmt.Errorf("ecies: unsupported ECDH algorithm")
	ErrUnsupportedECIESParameters = fmt.Errorf("ecies: unsupported ECIES parameters")
)
var paramsFromCurve = map[elliptic.Curve]*ECIESParams{
	elliptic.P256(): ECIES_AES128_SHA256,
	elliptic.P384(): ECIES_AES256_SHA384,
	elliptic.P521(): ECIES_AES256_SHA512,
}

//@overview: ecies 椭圆曲线集成加密方式
//@author: richard.sun
//@doc: https://cryptobook.nakov.com/asymmetric-key-ciphers/ecies-public-key-encryption
//@note:
//	ecc + kdf + symmetric encryption algorithm + mac
type ECIESParams struct {
	Hash      func() hash.Hash //hash function
	hashAlgo  crypto.Hash
	Cipher    func([]byte) (cipher.Block, error) // symmetric cipher
	BlockSize int                                // block size of symmetric cipher
	KeyLen    int                                // length of symmetric key
}

//@overview: 椭圆曲线公钥
type PubKey struct {
	X *big.Int
	Y *big.Int
	elliptic.Curve
	Params *ECIESParams
}

func NewECDSAPub(pub *ecdsa.PublicKey) *PubKey {
	return &PubKey{
		X:      pub.X,
		Y:      pub.Y,
		Curve:  pub.Curve,
		Params: paramsFromCurve[pub.Curve],
	}
}

func (pub *PubKey) ExportECDSA() *ecdsa.PublicKey {
	return &ecdsa.PublicKey{Curve: pub.Curve, X: pub.X, Y: pub.Y}
}

//@overview: 椭圆曲线私钥
type PriKey struct {
	Pub PubKey
	D   *big.Int
}

func NewECDSAPri(prv *ecdsa.PrivateKey) *PriKey {
	pub := NewECDSAPub(&prv.PublicKey)
	return &PriKey{*pub, prv.D}
}

func (prv *PriKey) ExportECDSA() *ecdsa.PrivateKey {
	pubECDSA := prv.Pub.ExportECDSA()
	return &ecdsa.PrivateKey{PublicKey: *pubECDSA, D: prv.D}
}

//@overview: 会话临时共享密钥
//@author: richard.sun
//@param:
//1. pub  客户端/发起端 公钥
func (prv *PriKey) SharedKey(pub *PubKey, skLen int, macLen int) ([]byte, error) {
	//1. 椭圆曲线算法 校验
	if prv.Pub.Curve != pub.Curve {
		return nil, ErrInvalidCurve
	}
	var length = utils.Max(skLen+macLen, (pub.Curve.Params().BitSize+7)>>3)
	var sk = make([]byte, length)
	//2. x 分量
	x, _ := pub.Curve.ScalarMult(pub.X, pub.Y, prv.D.Bytes())
	if x == nil {
		return nil, ErrSharedKeyIsPointAtInfinity
	}
	skBytes := x.Bytes()
	copy(sk[length-len(skBytes):], skBytes)
	return sk, nil
}

type IECIES interface {
	PriPub(rand io.Reader, curve elliptic.Curve) (*PriKey, error)
	Encrypt(rand io.Reader, pub *PubKey, m []byte, salt []byte) ([]byte, error)
	Decrypt(rand io.Reader, pri *PriKey, em []byte, salt []byte) ([]byte, error)
}

type ecies struct{}

func NewECIES() (IECIES, error) {
	return &ecies{}, nil
}

func (ec *ecies) PriPub(rand io.Reader, curve elliptic.Curve) (*PriKey, error) {
	var pb, x, y, err = elliptic.GenerateKey(curve, rand)
	if err != nil {
		return nil, err
	}
	return NewECDSAPri(&ecdsa.PrivateKey{ecdsa.PublicKey{curve, x, y}, new(big.Int).SetBytes(pb)}), nil
}

//@overview: 椭圆曲线公钥加密数据. Pub表示接收方的公钥.
// initiator 发送方 receiver 接收发
// iPri/iPub   发送方私钥/公钥
// rPri/rPub   接收方私钥/公钥
//对于发送方：具有iPri/iPub + rPub 信息. 加密逻辑是发送方生成临时密钥对，和rPub 生成临时共享密码, 然后
//使用对称加密算法加密数据. 数据格式
//
// iRandPub | cipher data | signature
//
//@author: richard.sun
//@param:
//1. rand  随机IO
//2. pub   接收方公钥
//3. m     待加密信息
//4. salt  盐
func (ec *ecies) Encrypt(rand io.Reader, pub *PubKey, m []byte, salt []byte) ([]byte, error) {
	//1. 获取公钥参数
	var (
		params = pub.Params
		randPP *PriKey
		sk     []byte
		err    error
		key    []byte
		em     []byte
	)
	if params == nil {
		return nil, ErrUnsupportedECDHAlgorithm
	}
	//2. 生成随机公私钥对
	randPP, err = ec.PriPub(rand, pub.Curve)
	if err != nil {
		return nil, err
	}
	//3. 生成临时共享密码对
	sk, err = randPP.SharedKey(pub, params.KeyLen, params.KeyLen)
	if err != nil {
		return nil, err
	}
	//4. KDF key
	key, err = ec.kdf(params.Hash, sk, salt, params.KeyLen)
	if err != nil {
		return nil, err
	}
	//5. 加密数据
	em, err = ec.encrypt(rand, params, key, m)
	if err != nil {
		return nil, err
	}
	//6. 数字签名
	mac := hmac.New(params.Hash, key)
	mac.Write(em)
	signature := mac.Sum(nil)
	//7. 构造完整数据
	randPub := elliptic.Marshal(pub.Curve, randPP.Pub.X, randPP.Pub.Y) //随机密钥中公钥
	ct := make([]byte, len(randPub)+len(em)+len(signature))
	copy(ct, randPub)
	copy(ct[len(randPub):], em)
	copy(ct[len(randPub)+len(em):], signature)

	return ct, nil
}

func (ec *ecies) encrypt(rand io.Reader, params *ECIESParams, key []byte, m []byte) ([]byte, error) {
	//1. 对称加密算法
	alg, err := params.Cipher(key)
	if err != nil {
		return nil, err
	}
	//2. iv 分量
	iv := make([]byte, params.BlockSize)
	_, err = io.ReadFull(rand, iv)
	if err != nil {
		return nil, err
	}
	//3. CTR
	ctr := cipher.NewCTR(alg, iv)
	//4. 加密数据并且携带IV数据信息
	ct := make([]byte, len(m)+params.BlockSize)
	copy(ct, iv)
	ctr.XORKeyStream(ct[params.BlockSize:], m)
	return ct, nil
}

func (ec *ecies) Decrypt(rand io.Reader, pri *PriKey, em []byte, salt []byte) ([]byte, error) {
	var params = pri.Pub.Params
	if params == nil {
		return nil, ErrUnsupportedECDHAlgorithm
	}
	var (
		h        = params.Hash()
		err      error
		iRandPub *PubKey
		sk       []byte
		key      []byte
		hashLen  = h.Size()
		pubLen   = (pri.Pub.Curve.Params().BitSize + 7) / 4
	)
	//1. 解析公钥
	iRandPub = new(PubKey)
	iRandPub.Curve = pri.Pub.Curve
	iRandPub.X, iRandPub.Y = elliptic.Unmarshal(iRandPub.Curve, em[:pubLen])
	if iRandPub.X == nil {
		return nil, ErrInvalidPublicKey
	}

	if !iRandPub.Curve.IsOnCurve(iRandPub.X, iRandPub.Y) {
		return nil, ErrInvalidCurve
	}
	//2. 会话临时共享密钥
	sk, err = pri.SharedKey(iRandPub, params.KeyLen, params.KeyLen)
	if err != nil {
		return nil, err
	}
	//3. KDF
	key, err = ec.kdf(params.Hash, sk, salt, params.KeyLen)
	if err != nil {
		return nil, err
	}
	//4. 生成签名
	mac := hmac.New(params.Hash, key)
	mac.Write(em[pubLen : len(em)-hashLen])
	signature := mac.Sum(nil)
	//5. 验证签名
	if !hmac.Equal(em[len(em)-hashLen:], signature) {
		return nil, ErrInvalidMessage
	}
	//6. 数据解密
	return ec.decrypt(params, key, em[pubLen:len(em)-hashLen])
}

func (ec *ecies) decrypt(params *ECIESParams, key []byte, em []byte) ([]byte, error) {
	alg, err := params.Cipher(key)
	if err != nil {
		return nil, err
	}
	ctr := cipher.NewCTR(alg, em[:params.BlockSize])
	m := make([]byte, len(em)-params.BlockSize)
	ctr.XORKeyStream(m, em[params.BlockSize:])
	return m, nil
}

//@overview: kdf is Key Derivation Functions.
//@author: richard.sun
//@param:
//1. h		 hash function
//2. secret  origin key
//3. salt
func (ec *ecies) kdf(h func() hash.Hash, secret []byte, salt []byte, keyLen int) ([]byte, error) {
	var (
		info   = []byte{0x6b, 0x65, 0x6c, 0x6c, 0x79, 0x63, 0x68, 0x65, 0x6e}
		r      = hkdf.New(h, secret, salt, info)
		key    = make([]byte, keyLen)
		_, err = io.ReadFull(r, key)
	)
	if err != nil {
		return nil, err
	}
	return key[:keyLen], nil
}
