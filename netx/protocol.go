package netx

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"time"
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
	//@overview: 协议连接关闭
	Done() <-chan struct{}
}

//@overview: 协议构造器
type ProtocolConstructor func(context.Context, net.Conn) IProtocol

//@overview: 协议处理器
type ProtocolHandler func(context.Context, []byte) []byte

//@overview: Stream
type frame []byte

func (m *frame) Bytes() []byte {
	return *m
}

type Stream struct {
	ws     chan frame
	rs     chan frame
	ec     chan error
	app    context.Context
	cancel context.CancelFunc
	quit   chan struct{}
	reader *bufio.Reader
	writer *bufio.Writer
	hLen   int
}

func NewStream(ctx context.Context, conn net.Conn) IProtocol {
	var s = &Stream{}
	s.ws = make(chan frame, 512)
	s.rs = make(chan frame, 512)
	s.ec = make(chan error, 128)
	s.app, s.cancel = context.WithCancel(ctx)
	s.reader = bufio.NewReader(conn)
	s.writer = bufio.NewWriter(conn)
	s.quit = make(chan struct{})
	s.hLen = 4

	go s.readLoop()
	go s.writeLoop()
	go s.errorLoop()

	return s
}

func (s *Stream) Done() <-chan struct{} {
	return s.quit
}

func (s *Stream) Unpacket(ctx context.Context) ([]byte, error) {
	select {
	case <-s.app.Done():
		return nil, s.app.Err()
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		var f = <-s.rs
		return f.Bytes(), nil
	}
}

func (s *Stream) Packet(ctx context.Context, body []byte) error {
	select {
	case <-s.app.Done():
		return s.app.Err()
	case <-ctx.Done():
		return ctx.Err()
	default:
		s.writePkt(body)
		return nil
	}
}

func (s *Stream) readLoop() {
	for {
		select {
		case <-s.app.Done():
			return
		default:
			b := s.read()
			s.rs <- frame(b)
		}
	}
}

func (s *Stream) writeLoop() {
	for v := range s.ws {
		s.writer.Write(v.Bytes())
		s.writer.Flush()
	}
}

func (s *Stream) read() []byte {
	var (
		bLen  int
		pLen  int
		body  []byte
		err   error
		delay = 1
	)
	for {
		switch {
		case (time.Duration(delay) * time.Millisecond) >= time.Second:
			s.ec <- io.EOF
		case err != nil:
			delay <<= 1
			time.Sleep(time.Duration(delay) * time.Millisecond)
		default:
			delay = 1
		}
		select {
		case <-s.app.Done():
			return nil
		default:
			bLen, err = s.readHeader()
			if err != nil {
				s.ec <- err
				continue
			}
			pLen = bLen + s.hLen //包总长度 = 包头长度 + 包体长度
			body, err = s.readPkt(pLen)
			if err != nil {
				s.ec <- err
				continue
			}
			if pLen <= s.hLen {
				continue
			}
			return body
		}
	}
}

func (s *Stream) readHeader() (int, error) {
	//1. 解析协议头部信息
	var b, err = s.reader.Peek(s.hLen)
	if err != nil {
		return 0, err
	}
	//2. 解析字节流 大端
	var bLen int32
	err = binary.Read(bytes.NewBuffer(b[:s.hLen]), binary.BigEndian, &bLen)
	if err != nil {
		return 0, err
	}
	//3. 返回包体长度
	return int(bLen), nil
}

func (s *Stream) readPkt(pLen int) ([]byte, error) {
	//1. 读取Header
	var (
		b   = make([]byte, pLen)
		n   int
		nn  int
		err error
	)
	n, err = s.reader.Read(b[:s.hLen])
	if err != nil || n < s.hLen {
		return nil, errReadPackage
	}
	//2. 读取Body
	for n < pLen {
		nn, err = s.reader.Read(b[n:])
		if err != nil {
			return nil, err
		}
		n += nn
	}
	//3. 解码网络字节流
	var body = make([]byte, pLen-s.hLen)
	err = binary.Read(bytes.NewBuffer(b[s.hLen:]), binary.BigEndian, body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (s *Stream) writeHeader(b *bytes.Buffer, bLen int) error {
	var (
		n   int
		h   = make([]byte, s.hLen)
		err = binary.Write(b, binary.BigEndian, int32(bLen))
	)
	if err != nil {
		return nil
	}
	copy(h, b.Bytes())
	b.Reset()
	n, err = b.Write(h)
	if n < s.hLen || err != nil {
		return errWritePackage
	}
	return err
}

func (s *Stream) writePkt(body []byte) {
	var (
		bLen = len(body)
		err  error
		b    = new(bytes.Buffer)
	)
	err = s.writeHeader(b, bLen)
	if err != nil {
		s.ec <- err
		return
	}
	err = binary.Write(b, binary.BigEndian, body)
	if err != nil {
		s.ec <- err
		return
	}
	s.ws <- frame(b.Bytes())
}

func (s *Stream) close() {
	s.cancel()
	s.quit <- struct{}{}
	time.Sleep(time.Second * 3)
}

func (s *Stream) errorLoop() {
	for e := range s.ec {
		switch {
		case e == io.EOF: //连接关闭
			s.close()
		}
	}
}
