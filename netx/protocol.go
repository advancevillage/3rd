package netx

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"net"
	"sync"
	"time"

	"github.com/advancevillage/3rd/utils"
)

var (
	errPartPackage  = errors.New("protocol part package") //粘包
	errParsePackage = errors.New("protocol parse package")
	errReadPackage  = errors.New("protocol read package error")
	errWritePackage = errors.New("protocol write package error")
)

type IProtocol interface {
	//@overview: 解包
	//@author: richard.sun
	//@note: 注意处理粘包问题 共享net.Conn时注意多goroutine并发读
	Unpacket(context.Context) ([]byte, error)
	//@overview: 封包
	//@author: richard.sun
	//@note: 注意处理拆包问题. 共享net.Conn时注意多goroutine并发写
	Packet(context.Context, []byte) error
}

//@overview: 协议构造器
type ProtocolConstructor func(net.Conn) IProtocol

//@overview: 协议处理器
type ProtocolHandler func(context.Context, []byte) ([]byte, error)

//@overview: HB = header + body
type HB struct {
	reader *bufio.Reader
	writer *bufio.Writer
	wmu    sync.Mutex
	rmu    sync.Mutex
	hLen   int
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
	p.rmu.Lock()
	defer p.rmu.Unlock()

	var bLen, err = p.readHeader(ctx)
	var b []byte
	if err != nil {
		return nil, err
	}
	switch {
	case p.reader.Size() < (bLen + p.hLen): //pkg size > buf size
		b = make([]byte, bLen+p.hLen)
		var bi = 0
		var bn = p.reader.Size()
		var n int
		for bi < (bLen + p.hLen) {
			n, err = p.reader.Read(b[bi : bi+bn])
			if err != nil || n < bn {
				return nil, errParsePackage
			}
			bi += bn
			bn = utils.Min(bLen+p.hLen-bi, bn)
			time.Sleep(50 * time.Millisecond)
		}
		var lb = make([]byte, bLen+p.hLen)
		err = binary.Read(bytes.NewBuffer(b), binary.BigEndian, lb)
		if err != nil {
			return nil, err
		}
		return lb[p.hLen:], nil
	case p.reader.Buffered() < (bLen + p.hLen): //pkg size <= buf size
		return nil, errPartPackage
	}
	//2. 拆包
	//example:  A/AB/A1A2B/AB1B2
	b, err = p.readBody(ctx, bLen+p.hLen)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (p *HB) Packet(ctx context.Context, body []byte) error {
	//1. 构造Header信息
	var h, err = p.writeHeader(ctx, body)
	if err != nil {
		return err
	}
	body, err = p.writeBody(ctx, body)
	if err != nil {
		return err
	}
	//2. 构造Body信息
	var pkg = new(bytes.Buffer)
	var n int
	n, err = pkg.Write(h)
	if err != nil {
		return err
	}
	if n < p.hLen {
		return errWritePackage
	}
	n, err = pkg.Write(body)
	if err != nil {
		return err
	}
	if n < len(body) {
		return errWritePackage
	}
	p.wmu.Lock()
	defer p.wmu.Unlock()

	n, err = p.writer.Write(pkg.Bytes())
	if err != nil {
		return err
	}
	if n < pkg.Len() {
		return errWritePackage
	}
	err = p.writer.Flush()
	if err != nil {
		return err
	}
	return nil
}

func (p *HB) readHeader(ctx context.Context) (int, error) {
	//1. 解析协议头部信息
	var b, err = p.reader.Peek(p.hLen)
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

func (p *HB) readBody(ctx context.Context, pl int) ([]byte, error) {
	var body = make([]byte, pl)
	var n, err = p.reader.Read(body)
	if err != nil || n < pl {
		return nil, errParsePackage
	}
	var t = make([]byte, pl)
	err = binary.Read(bytes.NewBuffer(body[:pl]), binary.BigEndian, t)
	if err != nil {
		return nil, err
	}
	return t[p.hLen:], nil
}

func (p *HB) writeHeader(ctx context.Context, body []byte) ([]byte, error) {
	var bLen = len(body)
	var b = new(bytes.Buffer)
	//1. 消息头
	var err = binary.Write(b, binary.BigEndian, int32(bLen))
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (p *HB) writeBody(ctx context.Context, body []byte) ([]byte, error) {
	var b = new(bytes.Buffer)
	var err = binary.Write(b, binary.BigEndian, body)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}
