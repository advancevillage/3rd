package netx

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
)

var (
	ErrPartPackage  = errors.New("protocol part package") //粘包
	ErrParsePackage = errors.New("protocol parse package")
	ErrReadPackage  = errors.New("protocol read package error")
	ErrWritePackage = errors.New("protocol write package error")
)

type IProtocol interface {
	//@overview: 解包
	//@author: richard.sun
	//@note: 注意处理粘包问题
	Unpacket(context.Context) ([]byte, error)
	//@overview: 封包
	//@author: richard.sun
	//@note: 注意处理拆包问题
	Packet(context.Context, []byte) error
	//@overview 错误订阅
	//@author: richard.sun
	HandleError(context.Context, error)
}

//@overview: 协议构造器
type ProtocolConstructor func(net.Conn) IProtocol

//@overview: 协议处理器
type ProtocolHandler func(context.Context, []byte) ([]byte, error)

//@overview: HB = header + body
type HB struct {
	hLen   int
	reader *bufio.Reader
	writer *bufio.Writer
}

func NewHBProtocol(conn net.Conn) IProtocol {
	return &HB{
		hLen:   4,
		reader: bufio.NewReader(conn),
		writer: bufio.NewWriter(conn),
	}
}

func (p *HB) Unpacket(ctx context.Context) ([]byte, error) {
	//1. 预分析协议
	var bLen, err = p.readHeader(ctx)
	var n int
	if err != nil {
		return nil, err
	}
	if p.reader.Buffered() < (bLen + p.hLen) {
		return nil, ErrPartPackage
	}
	//2. 拆包
	//example:  A/AB/A1A2B/AB1B2
	var body = make([]byte, bLen+p.hLen)
	n, err = p.reader.Read(body)
	if err != nil || n < (bLen+p.hLen) {
		return nil, ErrParsePackage
	}
	return body[p.hLen:], nil
}

func (p *HB) Packet(ctx context.Context, body []byte) error {
	//1. 构造Header信息
	var h, err = p.writeHeader(ctx, body)
	if err != nil {
		return err
	}
	//2. 构造Body信息
	var pkg = new(bytes.Buffer)
	var n int
	n, err = pkg.Write(h)
	if err != nil || n < p.hLen {
		return ErrWritePackage
	}
	n, err = pkg.Write(body)
	if err != nil || n < len(body) {
		return ErrWritePackage
	}
	n, err = p.writer.Write(pkg.Bytes())
	if err != nil || n < pkg.Len() {
		return ErrWritePackage
	}
	err = p.writer.Flush()
	if err != nil {
		return err
	}
	return nil
}

func (p *HB) HandleError(ctx context.Context, err error) {
	if err != nil {
		fmt.Println(err)
	}
}

func (p *HB) readHeader(ctx context.Context) (int, error) {
	//1. 解析协议头部信息
	var b, err = p.reader.Peek(int(p.hLen))
	if err != nil {
		return 0, err
	}
	//2. 解析字节流 大端
	var bLen int32
	err = binary.Read(bytes.NewBuffer(b[:p.hLen]), binary.BigEndian, &bLen)
	if err != nil {
		return 0, err
	}
	//3. 返回包长度
	return int(bLen), nil
}

func (p *HB) writeHeader(ctx context.Context, body []byte) ([]byte, error) {
	var bLen = len(body)
	var h = new(bytes.Buffer)
	//1. 消息头
	var err = binary.Write(h, binary.BigEndian, int32(bLen))
	if err != nil {
		return nil, err
	}
	return h.Bytes(), nil
}
