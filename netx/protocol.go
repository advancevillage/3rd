package netx

import (
	"bufio"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
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
	var ni = 0
	if err != nil {
		return nil, err
	}
	//2. 拆包
	var body = make([]byte, bLen+p.hLen)
	n, err = p.reader.Read(body[:p.hLen])
	if err != nil {
		return nil, errors.New("protocol parse error")
	}
	ni = n
	for ni < bLen {
		n, err = p.reader.Read(body[ni:])
		if err != nil {
			return nil, errors.New("protocol parse error")
		}
		ni += n
	}
	return body[p.hLen:], nil
}

func (p *HB) Packet(ctx context.Context, body []byte) error {
	var pkg, err = p.writeHeader(ctx, body)
	if err != nil {
		return err
	}
	var n int
	n, err = p.writer.Write(pkg)
	if err != nil || n < len(pkg) {
		return errors.New("protocol packet error")
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
	var x = binary.BigEndian.Uint32(b)
	//3. 返回包长度
	return int(x), nil
}

func (p *HB) writeHeader(ctx context.Context, body []byte) ([]byte, error) {
	var bLen = len(body)
	var pkg = make([]byte, p.hLen+bLen)
	binary.BigEndian.PutUint32(pkg[:p.hLen], uint32(bLen))
	copy(pkg[p.hLen:], body)
	return pkg, nil
}
