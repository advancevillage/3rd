package netx

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"fmt"
	"hash"
	"io"
)

var (
	zero16 = make([]byte, 16)
)

//@overview: 传输层接口协议. 特点是安全,灵活,高效
//@author: richard.sun
//@note:
//1. write frame
//2. read frame
type ITransport interface {
	Read(context.Context) ([]byte, error)
	Write(context.Context, []byte) error
}

type TransportOption struct {
	MaxSize uint32
}

type tcpConn struct {
}

//message auth code
//@overview: 加密通讯. Cipher Stream. 加密/解密
//@author: richard.sun
//@note:
//1. plain text  0 1 0 1 1 1 1 1 1 0 0 0 ..... 1 1 0 1 ...
//2. key stream  1 0 1 1 0 0 1 0 0 0 0 0 ..... 0 1 0 1 ...
//3. xor		 1 1 1 0 1 1 0 1 1 0 0 0 ..... 1 0 0 0 ...
type tcpMac struct {
	//数据流加密
	ens cipher.Stream
	des cipher.Stream
	//数据流签名
	macCipher cipher.Block
	egress    hash.Hash
	ingress   hash.Hash
}

//@overview: 读取加密帧. 基于自定义协议
//协议头 32bytes
//  [0 0 0 0][0 0 0 0][0 0 0 0 0 0 0 0]
//   帧长度4  标识位4    帧标识8
//  [0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 ]
//				帧签名16
//标识位:
//  1st bytes:  高4位 低4位 成对出现  ping=0x01  pong=0x10
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
func (m *tcpMac) readFrame(r io.Reader, hSize int, pad int) ([]byte, []byte, []byte, error) {
	var (
		header = make([]byte, hSize)
		err    error
	)
	//1. 读取头信息
	_, err = io.ReadFull(r, header)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("read frame header %s", err.Error())
	}
	//2. 验证头签名
	var expectHMac = m.hashMac(m.ingress, m.macCipher, header[:16])
	if !hmac.Equal(expectHMac, header[16:]) {
		return nil, nil, nil, fmt.Errorf("read frame bad header sign")
	}
	//3. 解密头信息
	m.des.XORKeyStream(header[:16], header[:16])
	var (
		frameSize = m.readInt32(header[:4])
		flags     = header[4:8]
		fId       = header[8:16]
		realSize  = int(frameSize)
		padding   = int(frameSize) % pad
	)
	if padding > 0 {
		realSize += 16 - padding
	}
	var frameBuf = make([]byte, realSize)
	_, err = io.ReadFull(r, frameBuf)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("read frame fId(%x) body  %s", fId, err.Error())
	}
	//4. 数据签名验证
	m.ingress.Write(frameBuf)
	var frameSeed = m.ingress.Sum(nil)
	expectHMac = m.hashMac(m.ingress, m.macCipher, frameSeed)

	_, err = io.ReadFull(r, header[16:32])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("read frame fId(%x) body sign %s", fId, err.Error())
	}
	if !hmac.Equal(expectHMac, header[16:]) {
		return nil, nil, nil, fmt.Errorf("read frame bad body sign")
	}
	//5. 解密数据体
	m.des.XORKeyStream(frameBuf[:frameSize], frameBuf[:frameSize])
	return flags, fId, frameBuf[:frameSize], nil
}

func (m *tcpMac) writeFrame(w io.Writer, hSize int, pad int, flags []byte, fId []byte, data []byte) error {
	//1. 创建头信息大小
	var (
		err       error
		header    = make([]byte, hSize)
		frameSize = len(data)
	)
	//2. 设置头信息中帧大小
	m.writeInt32(uint32(frameSize), header[:4])
	//3. 设置头信息中标识位
	copy(header[4:8], flags[:4])
	//4. 设置头信息中帧标识
	copy(header[8:16], fId[:8])
	//5. 加密头信息数据
	m.ens.XORKeyStream(header[:16], header[:16])
	//6. 头信息签名
	copy(header[16:], m.hashMac(m.egress, m.macCipher, header[:16]))
	_, err = w.Write(header)
	if err != nil {
		return fmt.Errorf("write frame fId(%x) header %s", fId, err.Error())
	}
	//7. 加密写入数据
	var mw = cipher.StreamWriter{S: m.ens, W: io.MultiWriter(w, m.egress)}
	_, err = mw.Write(data)
	if err != nil {
		return fmt.Errorf("write frame fId(%x) body %s", fId, err.Error())
	}
	//8. 补齐数据完整
	var padding = frameSize % pad
	switch {
	case padding > 0:
		_, err = mw.Write(zero16[:16-padding])
		if err != nil {
			return fmt.Errorf("write frame fId(%x) padding %s", fId, err.Error())
		}
	}
	//9. 数据签名
	var frameSeed = m.egress.Sum(nil)
	_, err = w.Write(m.hashMac(m.egress, m.macCipher, frameSeed))
	if err != nil {
		return fmt.Errorf("write frame fId(%x) body sign %s", fId, err.Error())
	}
	return nil
}

//@overview: 网络流大小端转换
func (m *tcpMac) readInt32(b []byte) uint32 {
	return uint32(b[3]) | uint32(b[2])<<8 | uint32(b[1])<<16 | uint32(b[0])<<24
}

func (m *tcpMac) writeInt32(v uint32, b []byte) {
	b[0] = byte(v >> 24)
	b[1] = byte(v >> 16)
	b[2] = byte(v >> 8)
	b[3] = byte(v)
}

func (m *tcpMac) hashMac(mac hash.Hash, block cipher.Block, seed []byte) []byte {
	var aesBuf = make([]byte, aes.BlockSize)
	block.Encrypt(aesBuf, mac.Sum(nil))
	for i := range aesBuf {
		aesBuf[i] ^= seed[i]
	}
	mac.Write(aesBuf)
	return mac.Sum(nil)[:16]
}
