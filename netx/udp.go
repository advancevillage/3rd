package netx

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/advancevillage/3rd/utils"
)

type IUDPServer interface {
	StartServer() error
}

var (
	udpPlainSrt = &Secrets{
		AK: []byte{
			0x52, 0x69, 0x63, 0x68, 0x61, 0x72, 0x64, 0x2e, 0x53, 0x75, 0x6e, 0x20, 0x6c, 0x6f, 0x76, 0x65,
			0x20, 0x4b, 0x65, 0x6c, 0x6c, 0x79, 0x2e, 0x43, 0x68, 0x65, 0x6e, 0x19, 0x95, 0x11, 0x01, 0xff,
		},
		MK: []byte{
			0x52, 0x69, 0x63, 0x68, 0x61, 0x72, 0x64, 0x2e, 0x53, 0x75, 0x6e, 0x20, 0x6c, 0x6f, 0x76, 0x65,
			0x20, 0x4b, 0x65, 0x6c, 0x6c, 0x79, 0x2e, 0x43, 0x68, 0x65, 0x6e, 0x19, 0x95, 0x11, 0x01, 0xff,
		},
		Egress:  sha256.New(),
		Ingress: sha256.New(),
	}
)

type udps struct {
	cfg     *ServerOption
	app     context.Context
	cancel  context.CancelFunc
	conn    *net.UDPConn
	cmpCli  utils.ICompress
	macCli  IMac
	handler Handler
	errChan chan error
	//读写队列
	rc chan *udpUnit
	wc chan *udpUnit
}

type udpUnit struct {
	msg
	addr *net.UDPAddr
}

func newUDPUnit(body []byte, addr *net.UDPAddr) *udpUnit {
	var u = &udpUnit{}
	u.data = body
	u.addr = addr
	return u
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
	s.rc = make(chan *udpUnit, 1024)
	s.wc = make(chan *udpUnit, 1024)
	s.handler = f
	s.cmpCli, err = utils.NewRLE()
	if err != nil {
		return nil, fmt.Errorf("create compress client fail. %s", err.Error())
	}
	s.macCli, err = NewMac(*udpPlainSrt)
	if err != nil {
		return nil, fmt.Errorf("create mac client fail. %s", err.Error())
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

	go s.readLoop()
	go s.writeLoop()

	for {
		select {
		case <-s.app.Done():
			return
		default:
			var buf = make([]byte, s.cfg.MaxSize)
			n, addr, err := s.conn.ReadFromUDP(buf)
			if err != nil {
				s.kelly(addr, err)
			} else {
				var t = time.NewTicker(time.Second)
				select {
				case s.rc <- newUDPUnit(buf[:n], addr):
				case <-t.C:
				}
			}
		}
	}
}

func (s *udps) readLoop() {
	var err error
	for m := range s.rc {
		if m == nil || m.addr == nil {
			continue
		}
		//1. 解密数据包
		m.flags, m.fId, m.data, err = s.macCli.ReadBytes(m.data, 0x20, 0x10)
		if err != nil {
			continue
		}
		go s.h(m)
	}
}

func (s *udps) writeLoop() {
	var err error
	for m := range s.wc {
		if m == nil || m.addr == nil || m.flags == nil || m.fId == nil {
			continue
		}
		//5. 加密
		m.data, err = s.macCli.WriteBytes(m.data, 0x20, 0x10, m.flags, m.fId)
		if err != nil {
			continue
		}
		s.conn.WriteToUDP(m.data, m.addr)
	}
}

func (s *udps) h(m *udpUnit) {
	var err error
	//2. 解压
	m.data, err = s.cmpCli.Uncompress(m.data)
	if err != nil {
		return
	}
	//3. 处理
	m.data = s.handler(s.app, m.data)
	//4. 压缩
	m.data, err = s.cmpCli.Compress(m.data)
	if err != nil {
		return
	}
	var t = time.NewTicker(time.Second)
	select {
	case s.wc <- m:
	case <-t.C:
	}
}

func (s *udps) kelly(addr *net.UDPAddr, err error) {
	//1. 压缩数据
	plain, _ := s.cmpCli.Compress([]byte(err.Error()))
	//2. 加密数据
	cipher, err := s.macCli.WriteBytes(plain, 0x20, 0x10, zero4, utils.UUID8Byte())
	if err != nil {
		return
	}
	s.conn.WriteToUDP(cipher, addr)
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
	cmpCli utils.ICompress
	macCli IMac
	notify chan struct{}

	rc chan *udpUnit
	wc chan *udpUnit
}

func NewUDPClient(cfg *ClientOption) (IUDPClient, error) {
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
	c.rc = make(chan *udpUnit, 1024)
	c.wc = make(chan *udpUnit, 1024)
	c.addr, err = net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", cfg.Host, cfg.UdpPort))
	if err != nil {
		return nil, err
	}
	c.cmpCli, err = utils.NewRLE()
	if err != nil {
		return nil, fmt.Errorf("create compress client fail. %s", err.Error())
	}
	c.macCli, err = NewMac(*udpPlainSrt)
	if err != nil {
		return nil, fmt.Errorf("create mac client fail. %s", err.Error())
	}
	c.conn, err = net.DialUDP("udp", nil, c.addr)
	if err != nil {
		return nil, err
	}

	go c.readLoop()
	go c.writeLoop()

	return c, nil
}

func (c *udpc) readLoop() {
	for {
		select {
		case <-c.app.Done():
		default:
			//1. 从网络IO读取数据
			var body = make([]byte, c.cfg.MaxSize)
			var t = time.NewTicker(time.Second)
			var u = &udpUnit{}
			n, err := c.conn.Read(body)
			if err != nil {
				u.err = err
			} else {
				u.flags, u.fId, u.data, u.err = c.macCli.ReadBytes(body[:n], 0x20, 0x10)
				u.data, _ = c.cmpCli.Uncompress(u.data)
			}
			select {
			case c.rc <- u:
			case <-t.C:
			}
		}

	}
}

func (c *udpc) writeLoop() {
	var err error
	for u := range c.wc {
		//1. 压缩数据
		u.data, err = c.cmpCli.Compress(u.data)
		if err != nil {
			continue
		}
		u.flags = zero4
		u.fId = utils.UUID8Byte()
		//2. 加密数据
		u.data, err = c.macCli.WriteBytes(u.data, 0x20, 0x10, u.flags, u.fId)
		if err != nil {
			continue
		}
		c.conn.Write(u.data)
	}
}

func (c *udpc) send(ctx context.Context, body []byte) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case c.wc <- newUDPUnit(body, nil):
		return nil
	}
}

func (c *udpc) receive(ctx context.Context) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case u := <-c.rc:
		return u.data, u.err
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
