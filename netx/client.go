package netx

import (
	"context"
	"errors"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type ITcpClient interface {
	Send(context.Context, []byte) error
	Receive(context.Context) ([]byte, error)
}

type ClientOpt struct {
	Address string              // 服务端地址 ip:port
	Timeout time.Duration       // 客户端超时
	Retry   int                 //重试次数
	PC      ProtocolConstructor //协议生成器
	MaxSize int                 //最大包
}

//@overview: Tcp协议客户端应该具备超时、重试、断开重连、发送请求的基本功能
type tcpClient struct {
	cfg    *ClientOpt
	p      IProtocol
	conn   net.Conn
	app    context.Context
	cancel context.CancelFunc
	mu     sync.Mutex
	notify chan struct{}
	wg     sync.WaitGroup
}

func NewTcpClient(cfg *ClientOpt) (ITcpClient, error) {
	//1. 参数检查
	if cfg == nil || len(cfg.Address) <= 0 || cfg.Timeout <= 0 || cfg.Retry <= 0 || cfg.PC == nil {
		return nil, errors.New("tcp client options invalid param")
	}
	//2. 构建客户端
	var c = &tcpClient{}
	c.cfg = cfg
	c.app, c.cancel = context.WithCancel(context.Background())
	c.notify = make(chan struct{})
	go c.loop()
	return c, nil
}

func (c *tcpClient) loop() {
	var err error
	go c.halfOpen()
	for {
		err = c.connect(c.app)
		if err != nil {
			return
		}
		select {
		case <-c.notify:
			c.conn.Close()
		case <-c.app.Done():
			c.conn.Close()
			return
		}
	}
}

func (c *tcpClient) connect(ctx context.Context) error {
	for {
		select {
		case <-c.app.Done():
			return nil
		default:
			var conn, err = net.Dial("tcp", c.cfg.Address)
			//1. 连接失败 重试连接
			if err != nil {
				time.Sleep(50 * time.Millisecond)
				continue
			}
			//2. 连接成功
			c.mu.Lock()
			c.conn = conn
			c.p = c.cfg.PC(c.app, c.conn)
			c.mu.Unlock()
			go c.heartbeat()
			return nil
		}
	}
}

func (c *tcpClient) reconnect() {
	c.notify <- struct{}{}
}

func (c *tcpClient) heartbeat() {
	defer c.reconnect()
	var err error
	for {
		//1. 参数校验
		select {
		case <-c.app.Done():
			return
		default:
			err = c.send(c.app, c.p, nil)
			if err != nil {
				return
			}
			time.Sleep(time.Second / 5)
		}
	}
}

func (c *tcpClient) send(ctx context.Context, p IProtocol, body []byte) error {
	var err error
	if p == nil {
		return errConnecting
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-p.Done():
		return errReconnected
	default:
		err = p.Packet(ctx, body)
		if err != nil {
			return err
		}
		return nil
	}
}

func (c *tcpClient) receive(ctx context.Context, p IProtocol) ([]byte, error) {
	var (
		b   []byte
		err error
	)
	if p == nil {
		return nil, errConnecting
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-p.Done():
		return nil, errReconnected
	default:
		b, err = p.Unpacket(ctx)
		if err != nil {
			return nil, err
		}
		return b, nil
	}
}

func (c *tcpClient) Send(ctx context.Context, body []byte) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-c.app.Done():
		return c.app.Err()
	case <-c.notify:
		return errReconnected
	default:
		return c.send(ctx, c.p, body)
	}
}

func (c *tcpClient) Receive(ctx context.Context) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-c.app.Done():
		return nil, c.app.Err()
	case <-c.notify:
		return nil, errReconnected
	default:
		return c.receive(ctx, c.p)
	}
}

func (c *tcpClient) halfOpen() {
	var pp = make(chan os.Signal)
	signal.Notify(pp, syscall.SIGPIPE)
	select {
	case <-c.app.Done():
	case <-pp:
		c.reconnect()
	}
}

func (c *tcpClient) close() {
	c.cancel()
	time.Sleep(time.Second)
	close(c.notify)
}

//udp client
type IUdpClient interface {
	Send(context.Context, []byte) error
	Receive(context.Context) ([]byte, error)
}

type udpClient struct {
	cfg    *ClientOpt
	conn   *net.UDPConn
	app    context.Context
	cancel context.CancelFunc
	addr   *net.UDPAddr
	notify chan struct{}
}

func NewUdpClient(cfg *ClientOpt) (IUdpClient, error) {
	//1. 参数检查
	if cfg == nil || len(cfg.Address) <= 0 {
		return nil, errors.New("udp client options invalid param")
	}
	//2. 构建客户端
	var c = &udpClient{}
	var err error
	c.cfg = cfg
	c.app, c.cancel = context.WithCancel(context.Background())
	c.notify = make(chan struct{})
	c.addr, err = net.ResolveUDPAddr("udp", cfg.Address)
	if err != nil {
		return nil, err
	}
	c.conn, err = net.DialUDP("udp", nil, c.addr)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (c *udpClient) send(ctx context.Context, body []byte) error {
	var err error
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		_, err = c.conn.Write(body)
		if err != nil {
			return err
		}
		return nil
	}
}

func (c *udpClient) receive(ctx context.Context) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		c.conn.SetDeadline(time.Now().Add(c.cfg.Timeout))
		var body = make([]byte, c.cfg.MaxSize)
		var n, err = c.conn.Read(body)
		if err != nil {
			return nil, err
		}
		return body[:n], nil
	}
}

func (c *udpClient) Send(ctx context.Context, body []byte) error {
	select {
	case <-c.app.Done():
		return c.app.Err()
	default:
		return c.send(ctx, body)
	}
}

func (c *udpClient) Receive(ctx context.Context) ([]byte, error) {
	select {
	case <-c.app.Done():
		return nil, c.app.Err()
	default:
		return c.receive(ctx)
	}
}
