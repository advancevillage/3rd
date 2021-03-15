package netx

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/advancevillage/3rd/utils"
)

type IUDPServer interface {
	StartServer() error
}

type udps struct {
	cfg     *ServerOption
	app     context.Context
	cancel  context.CancelFunc
	conn    *net.UDPConn
	cmpCli  utils.ICompress
	handler Handler
	errChan chan error
}

func NewUDPServer(cfg *ServerOption, f Handler) (IUDPServer, error) {
	//1. 参数校验
	if cfg == nil || len(cfg.Host) <= 0 || cfg.UdpPort < 0 || cfg.UdpPort > 65535 || f == nil {
		return nil, errors.New("opts param is invalid")
	}
	if cfg.MaxSize <= 0 {
		cfg.MaxSize = 1 << 16
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = time.Hour
	}
	var s = &udps{}
	var err error
	s.app, s.cancel = context.WithCancel(context.TODO())
	s.cfg = cfg
	s.errChan = make(chan error)
	s.cmpCli, err = utils.NewRLE()
	if err != nil {
		return nil, fmt.Errorf("create compress client fail. %s", err.Error())
	}
	return s, nil
}

func (s *udps) start() {
	var (
		err  error
		addr *net.UDPAddr
	)
	//1. 监听服务端口
	addr, err = net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.UdpPort))
	if err != nil {
		s.errChan <- err
		return
	}
	s.conn, err = net.ListenUDP("udp", addr)
	if err != nil {
		s.errChan <- err
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
				s.kelly(addr, err)
			} else {
				//3. 处理报文
				go s.h(addr, buf[:n])
			}
		}
	}
}

func (s *udps) h(addr *net.UDPAddr, body []byte) {
	var buf, _ = s.cmpCli.Uncompress(body)
	buf = s.handler(s.app, buf)
	buf, _ = s.cmpCli.Compress(buf)
	s.conn.WriteToUDP(buf, addr)
}

func (s *udps) kelly(addr *net.UDPAddr, err error) {
	buf, _ := s.cmpCli.Compress([]byte(err.Error()))
	s.conn.WriteToUDP(buf, addr)
}

func (s *udps) StartServer() error {
	go s.start()
	go waitQuitSignal(s.cancel)
	select {
	case <-s.app.Done():
		time.Sleep(time.Second)
		return nil
	case err := <-s.errChan:
		return err
	}
}

//////////////////////////////////////////////////////////////////
//udp client
type IUDPClient interface {
	Send(context.Context, []byte) error
	Receive(context.Context) ([]byte, error)
}

type udpc struct {
	cfg    *ClientOption
	conn   *net.UDPConn
	app    context.Context
	cancel context.CancelFunc
	addr   *net.UDPAddr
	notify chan struct{}
}

func NewUdpClient(cfg *ClientOption) (IUDPClient, error) {
	//1. 参数检查
	if cfg == nil || len(cfg.Host) <= 0 || cfg.UdpPort < 0 || cfg.UdpPort > 65535 {
		return nil, errors.New("opts param is invalid")
	}
	if cfg.MaxSize <= 0 {
		cfg.MaxSize = 1 << 16
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = time.Minute
	}
	//2. 构建客户端
	var c = &udpc{}
	var err error
	c.cfg = cfg
	c.app, c.cancel = context.WithCancel(context.Background())
	c.notify = make(chan struct{})
	c.addr, err = net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", cfg.Host, cfg.UdpPort))
	if err != nil {
		return nil, err
	}
	c.conn, err = net.DialUDP("udp", nil, c.addr)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (c *udpc) send(ctx context.Context, body []byte) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		n, err := c.conn.Write(body)
		if err != nil || n < len(body) {
			return fmt.Errorf("udp client write package %d. should be %d. %v", n, len(body), err)
		}
		return nil
	}
}

func (c *udpc) receive(ctx context.Context) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		var body = make([]byte, c.cfg.MaxSize)
		var n, err = c.conn.Read(body)
		if err != nil {
			return nil, err
		}
		return body[:n], nil
	}
}

func (c *udpc) Send(ctx context.Context, body []byte) error {
	if uint32(len(body)) > c.cfg.MaxSize {
		return fmt.Errorf("udp client write package %d over maxSize %d", len(body), c.cfg.MaxSize)
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return c.send(ctx, body)
	}
}

func (c *udpc) Receive(ctx context.Context) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return c.receive(ctx)
	}
}
