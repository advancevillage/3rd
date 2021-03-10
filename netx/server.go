package netx

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var (
	errConnectClosed = errors.New("connection closed")
	errConnected     = errors.New("connection connected")
	errConnecting    = errors.New("connection connecting")
	errReconnected   = errors.New("reconnection")
)

//@overview: 监听信号处理
func waitQuitSignal(cancel context.CancelFunc) {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c
	cancel()
}

type ServerOption struct {
	TransportOption
}

type ClientOption struct {
	TransportOption
	EnodeUrl string
}

///to delete
//tcp request handler func
type ITcpServer interface {
	StartServer()
	StopServer()
}

type ServerOpt struct {
	Host    string
	Port    int
	PC      ProtocolConstructor //协议生成器
	PH      ProtocolHandler     //协议处理器
	MaxSize int                 //最大报文长度
}

//@overview: tcp server. 目标是更多请求更少的内存消耗
//@author: richard.sun
type tcpServer struct {
	cfg    *ServerOpt
	app    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewTcpServer(cfg *ServerOpt) (ITcpServer, error) {
	//1. 参数校验
	if cfg == nil || cfg.Port < 0 || cfg.Port > 65535 {
		return nil, errors.New("opts param is invalid")
	}
	var s = &tcpServer{}
	s.app, s.cancel = context.WithCancel(context.Background())
	s.cfg = cfg

	return s, nil
}

func (s *tcpServer) start() {
	var (
		err      error
		delay    time.Duration
		protocol = "tcp"
		addr     = fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)
		l        net.Listener
		conn     net.Conn
	)
	//1. 监听端口
	l, err = net.Listen(protocol, addr)
	if err != nil {
		fmt.Printf("listen %s fail. %s\n", addr, err.Error())
		s.cancel()
		return
	}
	defer l.Close()
	//2. 监听请求
	for {
		select {
		case <-s.app.Done():
			return
		default:
			conn, err = l.Accept()
			//accept fail
			if err != nil {
				if ne, ok := err.(net.Error); ok && ne.Temporary() {
					switch {
					case delay < 5*time.Millisecond:
						delay = 5 * time.Millisecond
					case delay < time.Second:
						delay = 2 * delay
					case delay >= time.Second:
						delay = time.Second
					}
					var t = time.NewTicker(delay)
					select {
					case <-t.C:
					case <-s.app.Done():
						t.Stop()
						return
					}
				}
				continue
			}
			delay = 0
			//处理
			go s.handler(conn)
		}
	}
}

func (s *tcpServer) StartServer() {
	go s.start()
	go waitQuitSignal(s.cancel)
	select {
	case <-s.app.Done():
		time.Sleep(time.Second)
	}
}

func (s *tcpServer) StopServer() {
	s.cancel()
}

func (s *tcpServer) handler(conn net.Conn) {
	//1. 关闭链接
	s.wg.Add(1)
	defer s.wg.Done()
	defer conn.Close()
	//2. 上下文处理
	var ctx, cancel = context.WithCancel(s.app)
	defer cancel()
	//3. 协议编&解码器
	var p = s.cfg.PC(ctx, conn)
	var err error
	var body []byte

	for {
		select {
		case <-s.app.Done():
			return
		case <-p.Done():
			return
		default:
			//1. 解包
			body, err = p.Unpacket(ctx)
			if err != nil {
				continue
			}
			go s.handleFunc(ctx, p, body)
		}
	}
}

func (s *tcpServer) handleFunc(ctx context.Context, p IProtocol, body []byte) {
	s.wg.Add(1)
	defer s.wg.Done()
	var buf = s.cfg.PH(ctx, body)
	p.Packet(ctx, buf)
}

//udp server
type IUdpServer interface {
	StartServer()
	StopServer()
}

type udpServer struct {
	cfg    *ServerOpt
	app    context.Context
	cancel context.CancelFunc
	conn   *net.UDPConn
	wg     sync.WaitGroup
	ec     chan error
}

func NewUdpServer(cfg *ServerOpt) (IUdpServer, error) {
	//1. 检查参数
	if cfg == nil || nil == net.ParseIP(cfg.Host) || cfg.Port <= 0 || cfg.Port > 65535 {
		return nil, fmt.Errorf("invalid config param")
	}
	var u = &udpServer{}
	u.ec = make(chan error, 64)
	u.app, u.cancel = context.WithCancel(context.Background())
	u.cfg = cfg

	return u, nil
}

func (s *udpServer) StartServer() {
	go s.errorLoop()
	go s.start()
	go waitQuitSignal(s.cancel)
	select {
	case <-s.app.Done():
		close(s.ec)
		time.Sleep(time.Second * 3)
	}
}

func (s *udpServer) StopServer() {
	s.ec <- io.EOF
}

func (s *udpServer) start() {
	var (
		err  error
		addr *net.UDPAddr
	)
	//1. 监听服务端口
	addr, err = net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port))
	if err != nil {
		fmt.Println(err.Error())
		s.ec <- io.EOF
		return
	}
	s.conn, err = net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println(err.Error())
		s.ec <- io.EOF
		return
	}
	defer s.conn.Close()
	var (
		n   int
		buf = make([]byte, s.cfg.MaxSize)
	)
	for {
		select {
		case <-s.app.Done():
			return
		default:
			//2. 接收报文
			n, addr, err = s.conn.ReadFromUDP(buf)
			if err != nil {
				continue
			}
			var body = make([]byte, n)
			copy(body, buf[:n])
			//3. 处理报文
			go s.handle(addr, body)
		}
	}
}

func (s *udpServer) handle(addr *net.UDPAddr, body []byte) {
	s.wg.Add(1)
	defer s.wg.Done()
	var b = s.cfg.PH(s.app, body)
	s.conn.WriteToUDP(b, addr)
}

func (s *udpServer) errorLoop() {
	for e := range s.ec {
		switch {
		case e == io.EOF:
			s.cancel()
		}
	}
}
