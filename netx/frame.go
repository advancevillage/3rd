package netx

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"hash"
	"io"

	"github.com/advancevillage/3rd/ecies"
	"github.com/advancevillage/3rd/utils"
)

var (
	zero16    = make([]byte, 16)
	pingFlags = []byte{0x01, 0x0, 0x0, 0x0}
	pongFlags = []byte{0x10, 0x0, 0x0, 0x0}
)

type IMac interface {
	//IO
	ReadFrame(r io.Reader, hSize int, pad int) ([]byte, []byte, []byte, error)
	WriteFrame(w io.Writer, hSize int, pad int, flags []byte, fId []byte, data []byte) error
	//BytesBlock
	ReadBytes(cipher []byte, hSize int, pad int) ([]byte, []byte, []byte, error)
	WriteBytes(plain []byte, hSize int, pad int, flags []byte, fId []byte) ([]byte, error)
}

//@overview: 密钥信息
//@author: richard.sun
//@param:
// AK  aes.cipher 加密算法key 密钥流
// MK  aes.cipher 加密算法key 数据签名
// Egress  加密算法
// Ingress 加密算法
type Secrets struct {
	AK      []byte
	MK      []byte
	Egress  hash.Hash
	Ingress hash.Hash
}

//message auth code
//@overview: 加密通讯. Cipher Stream. 加密/解密
//@author: richard.sun
//@note:
//1. plain text  0 1 0 1 1 1 1 1 1 0 0 0 ..... 1 1 0 1 ...
//2. key stream  1 0 1 1 0 0 1 0 0 0 0 0 ..... 0 1 0 1 ...
//3. xor		 1 1 1 0 1 1 0 1 1 0 0 0 ..... 1 0 0 0 ...
type mac struct {
	//数据流加密
	ens cipher.Stream
	des cipher.Stream
	//数据流签名
	macCipher cipher.Block
	egress    hash.Hash
	ingress   hash.Hash
}

func NewMac(secret Secrets) (IMac, error) {
	var (
		macc cipher.Block
		encc cipher.Block
		err  error
	)
	macc, err = aes.NewCipher(secret.MK)
	if err != nil {
		return nil, err
	}
	encc, err = aes.NewCipher(secret.AK)
	if err != nil {
		return nil, err
	}
	var iv = make([]byte, encc.BlockSize())
	var m = &mac{
		ens:       cipher.NewCTR(encc, iv),
		des:       cipher.NewCTR(encc, iv),
		macCipher: macc,
		egress:    secret.Egress,
		ingress:   secret.Ingress,
	}
	return m, nil
}

//@overview: 读取加密帧. 基于自定义协议
//协议头 32bytes
//  [0 0 0 0][0 0 0 0][0 0 0 0 0 0 0 0]
//   帧长度4  标识位4    帧标识8
//  [0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 ]
//				帧签名16
//标识位:
//  1st bytes:  高4位 低4位 成对出现  ping=0x01  pong=0x10 (高4位 服务端 低4位 客户端)
//  2nd:  预留
//  3rd:  预留
//  4th:  预留
//@author: richard.sun
//@param:
//1. r		输入IO
//2. hSize	头部信息大小
//3. pad	补齐
//@return:
//1.  flags		[]byte  4
//2.  frame id	[]byte  4
//3.  data		[]byte
func (m *mac) ReadFrame(r io.Reader, hSize int, pad int) ([]byte, []byte, []byte, error) {
	var (
		header = make([]byte, hSize)
		err    error
		n      int
	)
	//1. 读取头信息
	n, err = io.ReadFull(r, header)
	if err != nil || n < hSize {
		return nil, nil, nil, fmt.Errorf("read frame header fail. %d < %d", n, hSize)
	}
	//2. 验证头签名
	var expectHMac = m.hashMac(m.ingress, m.macCipher, header[:0x10])
	if !hmac.Equal(expectHMac, header[0x10:]) {
		return nil, nil, nil, fmt.Errorf("read frame bad header sign %x != %x", expectHMac, header[0x10:])
	}
	//3. 解密头信息
	m.des.XORKeyStream(header[:0x10], header[:0x10])
	var (
		frameSize = m.readInt32(header[:0x04])
		flags     = header[0x04:0x08]
		fId       = header[0x08:0x10]
		realSize  = int(frameSize)
		padding   = int(frameSize) % pad
	)
	if padding > 0 {
		realSize += 0x10 - padding
	}
	var frameBuf = make([]byte, realSize)
	n, err = io.ReadFull(r, frameBuf)
	if err != nil || n < realSize {
		return nil, nil, nil, fmt.Errorf("read frame fId(%x) body fail. %d < %d. %v", fId, n, realSize, err)
	}
	//4. 数据签名验证
	m.ingress.Write(frameBuf)
	var frameSeed = m.ingress.Sum(nil)
	expectHMac = m.hashMac(m.ingress, m.macCipher, frameSeed)

	n, err = io.ReadFull(r, header[0x10:0x20])
	if err != nil || n < 0x10 {
		return nil, nil, nil, fmt.Errorf("read frame fId(%x) body sign. %d < %d. %v", fId, n, 0x10, err)
	}
	if !hmac.Equal(expectHMac, header[0x10:]) {
		return nil, nil, nil, fmt.Errorf("read frame bad body sign. %x != %x", expectHMac, header[0x10:0x20])
	}
	//5. 解密数据体
	m.des.XORKeyStream(frameBuf, frameBuf)
	return flags, fId, frameBuf[:frameSize], nil
}

func (m *mac) WriteFrame(w io.Writer, hSize int, pad int, flags []byte, fId []byte, data []byte) error {
	//1. 创建头信息大小
	var (
		n         int
		err       error
		header    = make([]byte, hSize)
		frameSize = len(data)
	)
	//2. 设置头信息中帧大小
	m.writeInt32(uint32(frameSize), header[:0x04])
	//3. 设置头信息中标识位
	copy(header[0x04:0x08], flags[:0x04])
	//4. 设置头信息中帧标识
	copy(header[0x08:0x10], fId[:0x08])
	//5. 加密头信息数据
	m.ens.XORKeyStream(header[:0x10], header[:0x10])
	//6. 头信息签名
	copy(header[0x10:], m.hashMac(m.egress, m.macCipher, header[:0x10]))
	n, err = w.Write(header)
	if err != nil || n < hSize {
		return fmt.Errorf("write frame fId(%x) header fail. %d < %d. %v", fId, n, hSize, err)
	}
	//7. 加密写入数据
	var mw = cipher.StreamWriter{S: m.ens, W: io.MultiWriter(w, m.egress)}
	n, err = mw.Write(data)
	if err != nil || n < len(data) {
		return fmt.Errorf("write frame fId(%x) body fail. %d < %d. %v", fId, n, len(data), err)
	}
	//8. 补齐数据完整
	var padding = frameSize % pad
	switch {
	case padding > 0:
		n, err = mw.Write(zero16[:0x10-padding])
		if err != nil || n < 0x10-padding {
			return fmt.Errorf("write frame fId(%x) padding fail. %d < %d. %v", fId, n, 0x10-padding, err)
		}
	}
	//9. 数据签名
	var frameSeed = m.egress.Sum(nil)
	var endSign = m.hashMac(m.egress, m.macCipher, frameSeed)
	n, err = w.Write(endSign)
	if err != nil || n < len(endSign) {
		return fmt.Errorf("write frame fId(%x) body sign fail. %d < %d. %v", fId, n, len(endSign), err)
	}
	return nil
}

func (m *mac) ReadBytes(cipher []byte, hSize int, pad int) ([]byte, []byte, []byte, error) {
	var ingress = sha256.New()
	//1. 读取头信息
	if len(cipher) < hSize {
		return nil, nil, nil, fmt.Errorf("read bytes header error %d < %d", len(cipher), hSize)
	}
	var header = cipher[:hSize]
	//2. 验证头签名
	var expectHMac = m.hashMac(ingress, m.macCipher, header[:0x10])
	if !hmac.Equal(expectHMac, header[0x10:]) {
		return nil, nil, nil, fmt.Errorf("read bytes bad header sign %x != %x", expectHMac, header[0x10:])
	}
	//3. 解析头信息
	var (
		frameSize = m.readInt32(header[:0x4])
		flags     = header[0x4:0x8]
		fId       = header[0x8:0x10]
		realSize  = int(frameSize)
		padding   = int(frameSize) % pad
	)
	if padding > 0 {
		realSize += 16 - padding
	}
	//3.1 截取加密体
	cipher = cipher[hSize:]
	if len(cipher) < realSize {
		return nil, nil, nil, fmt.Errorf("read bytes fId(%x) body error. %d < %d", fId, len(cipher), realSize)
	}
	var frameBuf = cipher[:realSize]
	//4. 内容待解密
	ingress.Write(frameBuf)
	var frameSeed = ingress.Sum(nil)
	expectHMac = m.hashMac(ingress, m.macCipher, frameSeed)
	//5. 验证内容签名
	cipher = cipher[realSize:]
	if len(cipher) < 0x10 {
		return nil, nil, nil, fmt.Errorf("read bytes end sign fail. %d < %d", len(cipher), 0x10)
	}
	if !hmac.Equal(expectHMac, cipher[:0x10]) {
		return nil, nil, nil, fmt.Errorf("read bytes bad body sign %x != %x", expectHMac, cipher[:0x10])
	}
	return flags, fId, frameBuf[:frameSize], nil
}

func (m *mac) WriteBytes(plain []byte, hSize int, pad int, flags []byte, fId []byte) ([]byte, error) {
	//1. 创建头信息大小
	var (
		err       error
		header    = make([]byte, hSize)
		frameSize = len(plain)
		b         = bytes.NewBuffer([]byte{})
		n         int
		egress    = sha256.New()
	)
	//2. 设置头信息中帧大小
	m.writeInt32(uint32(frameSize), header[:4])
	//3. 设置头信息中标识位
	copy(header[4:8], flags[:4])
	//4. 设置头信息中帧标识
	copy(header[8:16], fId[:8])
	//5. 头信息签名
	copy(header[16:], m.hashMac(egress, m.macCipher, header[:16]))
	n, err = b.Write(header)
	if err != nil || n < hSize {
		return nil, fmt.Errorf("write bytes fId(%x) header %v", fId, err)
	}
	//7. 写入数据
	var mw = io.MultiWriter(b, egress)
	n, err = mw.Write(plain)
	if err != nil || n < len(plain) {
		return nil, fmt.Errorf("write bytes fId(%x) err %v", fId, err)
	}
	//8. 补齐数据完整
	var padding = frameSize % pad
	switch {
	case padding > 0:
		_, err = mw.Write(zero16[:16-padding])
		if err != nil {
			return nil, fmt.Errorf("write bytes fId(%x) padding %s", fId, err.Error())
		}
	}
	//9. 数据签名
	var frameSeed = egress.Sum(nil)
	var endSign = m.hashMac(egress, m.macCipher, frameSeed)
	n, err = b.Write(endSign)
	if err != nil || n < len(endSign) {
		return nil, fmt.Errorf("write bytes fId(%x) body sign fail. %d < %d. %v", fId, n, len(endSign), err)
	}
	return b.Bytes(), nil
}

//@overview: 网络流大小端转换
func (m *mac) readInt32(b []byte) uint32 {
	return uint32(b[3]) | uint32(b[2])<<8 | uint32(b[1])<<16 | uint32(b[0])<<24
}

func (m *mac) writeInt32(v uint32, b []byte) {
	b[0] = byte(v >> 24)
	b[1] = byte(v >> 16)
	b[2] = byte(v >> 8)
	b[3] = byte(v)
}

func (m *mac) hashMac(mac hash.Hash, block cipher.Block, seed []byte) []byte {
	var aesBuf = make([]byte, aes.BlockSize)
	block.Encrypt(aesBuf, mac.Sum(nil))
	for i := range aesBuf {
		aesBuf[i] ^= seed[i]
	}
	mac.Write(aesBuf)
	return mac.Sum(nil)[:16]
}

//@overview: 握手协议, 建立TCP连接后发送的第一个报文
//@author: richard.sun
//@param:
//1. ecies        密钥交换握手加密
//2. self enode   服务节点通讯信息
//3. remote enode 客户端节点 临时存储发送端公钥
type ecdhe struct {
	ecies     ecies.IECIES
	self      ecies.IENode
	ephemeral Secrets
}

type IECDHE interface {
	Write(w io.Writer, pub *ecdsa.PublicKey) (*ecdsa.PrivateKey, []byte, error)
	Read(r io.Reader) (*ecdsa.PublicKey, *ecdsa.PublicKey, []byte, error)
	Ephemeral(iRandPri *ecdsa.PrivateKey, inonce []byte, rRandPub *ecdsa.PublicKey, rnonce []byte) (*Secrets, error)
}

func NewECDHE256(pri *ecdsa.PrivateKey, host string, tcpPort int, udpPort int) (IECDHE, error) {
	self, err := ecies.NewENodeByPP(pri, host, tcpPort, udpPort)
	if err != nil {
		return nil, fmt.Errorf("new ecdhe 256 fail because of %s", err.Error())
	}
	ecies, err := ecies.NewECIES()
	if err != nil {
		return nil, fmt.Errorf("new ecdhe 256 fail because of %s", err.Error())
	}
	return &ecdhe{
		ecies: ecies,
		self:  self,
	}, nil
}

//@author: richard.sun
//@param:
//1. pub  接收方公钥
func (hs *ecdhe) Write(w io.Writer, pub *ecdsa.PublicKey) (*ecdsa.PrivateKey, []byte, error) {
	//1. 构造参数
	var (
		n      int
		sk     []byte
		token  []byte
		err    error
		randPP *ecdsa.PrivateKey
		nonce  = make([]byte, 0x20)
		pri    = ecies.NewECDSAPri(hs.self.GetPriKey())
	)
	//2. 生成随机数
	n, err = rand.Read(nonce)
	if err != nil || n < 0x20 {
		return nil, nil, fmt.Errorf("ecdhe write nonce fail. %v", err)
	}
	//3. 生成随机密钥
	randPP, err = ecdsa.GenerateKey(hs.self.GetPubKey().Curve, rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("ecdhe write random ecc private key fail. %s", err.Error())
	}
	//4. 生成共享临时密钥
	sk, err = pri.SharedKey(ecies.NewECDSAPub(pub), 0x10, 0x10)
	if err != nil {
		return nil, nil, fmt.Errorf("ecdhe write random share key fail. %s", err.Error())
	}
	token = utils.Xor(sk, nonce)
	//5. 签名
	var h = hmac.New(pri.Pub.Params.Hash, make([]byte, pri.Pub.Params.BlockSize))
	h.Write(token)
	h.Write(elliptic.Marshal(randPP.PublicKey.Curve, randPP.PublicKey.X, randPP.PublicKey.Y))
	var signature = h.Sum(nil)
	//6. 构造数据
	var data = make([]byte, 0xc2)
	copy(data[:0x41], elliptic.Marshal(hs.self.GetPubKey().Curve, hs.self.GetPubKey().X, hs.self.GetPubKey().Y))
	copy(data[0x41:0x61], nonce)
	copy(data[0x61:0x81], signature)
	copy(data[0x81:0xc2], elliptic.Marshal(randPP.PublicKey.Curve, randPP.PublicKey.X, randPP.PublicKey.Y))
	//7. 加密数据
	em, err := hs.ecies.Encrypt(rand.Reader, ecies.NewECDSAPub(pub), data, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("ecdhe encrypt data fail. %s", err.Error())
	}
	n, err = w.Write(em)
	if err != nil || n < len(em) {
		return nil, nil, fmt.Errorf("ecdhe send encrypt data fail. %v", err)
	}
	return randPP, nonce, nil
}

//@overview: 握手协议报文 格式. 基于secp256r1 加密方式
//
// iRandPub | iv | chiper data | signature
// signature = sha256(iv | chiper data)
//
// len(iRandPub) = 65  secp256r1
// len(iv)		 = 16
// len(signature)= 32  SHA256
//
// chiper data :  iPub | nonce | signature | ePub
//				   65      32       32        65
// token  = ecc(iPub, rPri)
// signed = token ^ nonce
// signature = sha256(signed, ePub)
//
func (hs *ecdhe) Read(r io.Reader) (*ecdsa.PublicKey, *ecdsa.PublicKey, []byte, error) {
	//1. 构造本地私钥用于解密
	var (
		pri = ecies.NewECDSAPri(hs.self.GetPriKey())
		err error
		em  = make([]byte, 0x133)
		m   []byte
		n   int
	)
	//2. 读取数据
	n, err = io.ReadFull(r, em)
	if err != nil || n < 0x133 {
		return nil, nil, nil, fmt.Errorf("ecdhe handshake fail due to having not enough data. %v", err)
	}
	//3. 解密数据
	m, err = hs.ecies.Decrypt(rand.Reader, pri, em, nil)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("ecdhe handshake fail due to decrypt em data. %s", err.Error())
	}
	//4. 解析数据
	var (
		iPubBytes = m[:0x41]
		nonce     = m[0x41:0x61]
		signature = m[0x61:0x81]
		ePubBytes = m[0x81:0xc2]
		x, y      = elliptic.Unmarshal(hs.self.GetPubKey().Curve, iPubBytes)
		xx, yy    = elliptic.Unmarshal(hs.self.GetPubKey().Curve, ePubBytes)
		iPub      = &ecdsa.PublicKey{hs.self.GetPubKey().Curve, x, y}
		ePub      = &ecdsa.PublicKey{hs.self.GetPubKey().Curve, xx, yy}
		token     []byte
	)
	//5. token
	sk, err := pri.SharedKey(ecies.NewECDSAPub(iPub), 0x10, 0x10)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("ecdhe handshake parse data fail. %s", err.Error())
	}
	token = utils.Xor(sk, nonce)
	//6. 校验签名
	var h = hmac.New(pri.Pub.Params.Hash, make([]byte, pri.Pub.Params.KeyLen))
	h.Write(token)
	h.Write(ePubBytes)
	var should = h.Sum(nil)
	if !hmac.Equal(should, signature) {
		return nil, nil, nil, fmt.Errorf("ecdhe handshake auth signature fail.")
	}
	//7. 获取临时会话通讯ePub
	return iPub, ePub, nonce, nil
}

//@overview: 握手协议交换密钥生成临时密钥
//@author: richard.sun
func (hs *ecdhe) Ephemeral(iRandPri *ecdsa.PrivateKey, inonce []byte, rRandPub *ecdsa.PublicKey, rnonce []byte) (*Secrets, error) {
	//1. 参数校验
	if iRandPri == nil || rRandPub == nil || len(inonce) <= 0 || len(rnonce) <= 0 || len(inonce) != len(rnonce) {
		return nil, fmt.Errorf("ecdhe init secret fail due to invalid param")
	}
	//2. 初始化临时密钥对
	var (
		err error
		sk  []byte
		h   = sha256.New()
	)
	sk, err = ecies.NewECDSAPri(iRandPri).SharedKey(ecies.NewECDSAPub(rRandPub), 0x10, 0x10)
	if err != nil {
		return nil, fmt.Errorf("ecdhe init secret fail due to share key %s", err.Error())
	}
	h.Write(sk)
	h.Write(utils.Xor(inonce, rnonce))
	hs.ephemeral.AK = h.Sum(nil)
	h.Write(hs.ephemeral.AK)
	hs.ephemeral.MK = h.Sum(nil)
	hs.ephemeral.Egress = sha256.New()
	hs.ephemeral.Ingress = sha256.New()
	return &hs.ephemeral, nil
}
