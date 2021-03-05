package netx

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/advancevillage/3rd/utils"
)

type ServerOption struct {
	TransportOption
}

type ITCPServer interface {
	StartServer()
}

type Handler func(context.Context, []byte) []byte

type server struct {
	cmpCli utils.ICompress
	heCli  IECDHE
	app    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	//cfg
	cfg     *ServerOption
	handler Handler
}

func NewTCPServer(cfg *ServerOption, f Handler) (ITCPServer, error) {
	//1. 参数检查
	if cfg == nil || len(cfg.Host) <= 0 || cfg.Port > 65535 || cfg.Port <= 0 || cfg.UdpPort > 65535 || cfg.UdpPort <= 0 || f == nil {
		return nil, fmt.Errorf("tcp invalid config param")
	}
	if cfg.MaxSize <= 0 {
		cfg.MaxSize = 1 << 16
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = time.Second * 10
	}
	var ctx, cancel = context.WithCancel(context.TODO())
	var err error
	var s = &server{
		app:     ctx,
		cancel:  cancel,
		cfg:     cfg,
		handler: f,
	}
	s.cmpCli, err = utils.NewRLE()
	if err != nil {
		return nil, fmt.Errorf("create compress client fail. %s", err.Error())
	}
	s.heCli, err = NewECDHE256(cfg.PriKey, cfg.Host, cfg.Port, cfg.UdpPort)
	if err != nil {
		return nil, fmt.Errorf("create ecdhe client fail. %s", err.Error())
	}
	return s, nil
}

func (s *server) start() {
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
			go s.dealwith(conn)
		}
	}
}

func (s *server) dealwith(conn net.Conn) {
	defer conn.Close()
	//1. 握手协商密钥 第一个包
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	clientPub, clientEphemeralPub, clientEphemeralNonce, err := s.heCli.Read(conn)
	if err != nil {
		conn.Write([]byte(err.Error()))
		time.Sleep(time.Second)
		return
	}
	serverEphemeralPub, serverEphemeralNonce, err := s.heCli.Write(conn, clientPub)
	if err != nil {
		conn.Write([]byte(err.Error()))
		time.Sleep(time.Second)
		return
	}
	//2. 生成临时密钥对
	srt, err := s.heCli.Ephemeral(serverEphemeralPub, serverEphemeralNonce, clientEphemeralPub, clientEphemeralNonce)
	if err != nil {
		conn.Write([]byte(err.Error()))
		time.Sleep(time.Second)
		return
	}
	//3. 构建加密传输通道
	cc, err := NewConn(conn, &s.cfg.TransportOption, srt)
	if err != nil {
		conn.Write([]byte(err.Error()))
		time.Sleep(time.Second)
		return
	}
	conn.SetDeadline(time.Time{})
	for {
		select {
		case <-s.app.Done():
			return
		default:
			//4. 读取数据流
			buf, err := cc.Read(s.app)
			if err != nil {
				conn.Write([]byte(err.Error()))
				time.Sleep(time.Second)
				return
			}
			go s.h(cc, buf)
		}
	}
}

func (s *server) h(tsp ITransport, data []byte) {
	var buf, _ = s.cmpCli.Uncompress(data)
	buf = s.handler(s.app, buf)
	buf, _ = s.cmpCli.Compress(buf)
	tsp.Write(s.app, buf)
}

func (s *server) StartServer() {
	go s.start()
	go waitQuitSignal(s.cancel)
	select {
	case <-s.app.Done():
		time.Sleep(time.Second)
	}
}
