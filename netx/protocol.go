package netx

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"log"
	"net"
	"sync"

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
	//@overview 错误订阅
	//@author: richard.sun
	HandleError(context.Context, error)
}

//@overview: 协议构造器
type ProtocolConstructor func(net.Conn, *TcpProtocolOpt) IProtocol

//@overview: 协议处理器
type ProtocolHandler func(context.Context, []byte) ([]byte, error)

//@overview: HB = header + body
type HB struct {
	reader *bufio.Reader
	writer *bufio.Writer
	wmu    sync.Mutex
	rmu    sync.Mutex
	cfg    *TcpProtocolOpt
}

type TcpProtocolOpt struct {
	MP IMultiPlexer
}

func NewHBProtocol(conn net.Conn, cfg *TcpProtocolOpt) IProtocol {
	return &HB{
		cfg:    cfg,
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
	if p.reader.Buffered() < (bLen + p.cfg.MP.Size()) {
		return nil, errPartPackage
	}
	//2. 拆包
	//example:  A/AB/A1A2B/AB1B2
	b, err = p.readBody(ctx, bLen+p.cfg.MP.Size())
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
	if n < p.cfg.MP.Size() {
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

func (p *HB) HandleError(ctx context.Context, err error) {
	if err != nil {
		log.Println(err)
	}
}

func (p *HB) readHeader(ctx context.Context) (int, error) {
	//1. 解析协议头部信息
	var b, err = p.reader.Peek(p.cfg.MP.Size())
	if err != nil {
		return 0, err
	}
	//2. 解析字节流 大端
	var h = make([]byte, p.cfg.MP.Size())
	var hs []byte
	var bLen int32
	err = binary.Read(bytes.NewBuffer(b[:p.cfg.MP.Size()]), binary.BigEndian, h)
	if err != nil {
		return 0, err
	}
	hs, err = p.cfg.MP.Demulti(h)
	if err != nil {
		return 0, err
	}
	err = binary.Read(bytes.NewBuffer(hs), binary.LittleEndian, &bLen)
	if err != nil {

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
	return t[p.cfg.MP.Size():], nil
}

func (p *HB) writeHeader(ctx context.Context, body []byte) ([]byte, error) {
	var bLen = len(body)
	var b = new(bytes.Buffer)
	//1. 消息头
	var err = binary.Write(b, binary.LittleEndian, int32(bLen))
	if err != nil {
		return nil, err
	}
	var h []byte
	h, err = p.cfg.MP.Multi(b.Bytes())
	if err != nil {
		return nil, err
	}
	b.Reset()
	err = binary.Write(b, binary.BigEndian, h)
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

//@overview: multiplexer. 复用器. 共享net.Conn
type IMultiPlexer interface {
	Demulti(h []byte) ([]byte, error)
	Multi(h []byte) ([]byte, error)
	Size() int
}

type mp struct {
	id []byte
	bi int //边界索引
	hs int //头信息长度
}

func NewMultiPlexer(hs int, bi int) IMultiPlexer {
	return &mp{
		id: make([]byte, hs-bi),
		hs: hs,
		bi: bi,
	}
}

func (m *mp) Size() int {
	return m.hs
}

func (m *mp) Demulti(h []byte) ([]byte, error) {
	if len(h) != m.hs {
		return nil, errParsePackage
	}
	copy(m.id, h[m.bi:])
	log.Println("demulti", string(h), string(m.id), string(h[:m.bi]))
	h = h[:m.bi]
	return h, nil
}

func (m *mp) Multi(b []byte) ([]byte, error) {
	var h = make([]byte, m.hs)
	copy(h[m.bi:], utils.SnowFlakeIdBytes(m.hs-m.bi))
	copy(h[:m.bi], b)
	log.Println("multi", string(b), string(h[m.bi:]), string(h))
	return h, nil
}
