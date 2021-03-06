package netx

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/advancevillage/3rd/ecies"
	"github.com/advancevillage/3rd/utils"
)

type ITCPServer interface {
	StartServer() error
}

type tcps struct {
	cmpCli  utils.ICompress
	heCli   IECDHE
	app     context.Context
	cancel  context.CancelFunc
	errChan chan error

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
		cfg.Timeout = time.Hour
	}
	var ctx, cancel = context.WithCancel(context.TODO())
	var err error
	var s = &tcps{
		app:     ctx,
		cancel:  cancel,
		cfg:     cfg,
		handler: f,
		errChan: make(chan error),
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

func (s *tcps) start() {
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
		s.errChan <- fmt.Errorf("listen %s fail. %s\n", addr, err.Error())
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

func (s *tcps) dealwith(conn net.Conn) {
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
	//3. 注意 握手成功后发送的数据均被压缩传输
	for {
		select {
		case <-s.app.Done():
			return
		default:
			//4. 读取数据流
			m, err := cc.Read(s.app)
			if err != nil {
				s.kelly(cc, err)
				return
			}
			go s.h(cc, m)
		}
	}
}

func (s *tcps) h(tsp ITransport, m *msg) {
	var buf, _ = s.cmpCli.Uncompress(m.data)
	buf = s.handler(s.app, buf)
	buf, _ = s.cmpCli.Compress(buf)
	m.data = buf
	tsp.Write(s.app, m)
}

func (s *tcps) kelly(tsp ITransport, err error) {
	buf, _ := s.cmpCli.Compress([]byte(err.Error()))
	tsp.Write(s.app, newmsg(buf))
	time.Sleep(time.Second)
}

func (s *tcps) StartServer() error {
	go s.start()
	go waitQuitSignal(s.cancel)
	select {
	case <-s.app.Done():
		time.Sleep(time.Second)
		return nil
	case e := <-s.errChan:
		return e
	}
}

//////////////////////////////////////////////////////////////////////////////////////
//tcp client
type ITCPClient interface {
	Send(context.Context, []byte) error
	Receive(context.Context) ([]byte, error)
}

type tcpc struct {
	cmpCli utils.ICompress
	heCli  IECDHE
	app    context.Context
	cancel context.CancelFunc
	conn   ITransport
	notify chan struct{}
	err    error

	//cfg
	cfg *ClientOption
	svr ecies.IENode
}

func NewTCPClient(cfg *ClientOption) (ITCPClient, error) {
	//1. 参数检查
	if cfg == nil || cfg.PriKey == nil || len(cfg.EnodeUrl) <= 0 {
		return nil, fmt.Errorf("tcp invalid config param")
	}
	if cfg.MaxSize <= 0 {
		cfg.MaxSize = 1 << 16
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = time.Second * 30
	}
	//2. 构造客户端
	var ctx, cancel = context.WithCancel(context.TODO())
	var err error
	var c = &tcpc{
		app:    ctx,
		cancel: cancel,
		cfg:    cfg,
	}
	c.cmpCli, err = utils.NewRLE()
	if err != nil {
		return nil, fmt.Errorf("create compress client fail. %s", err.Error())
	}
	c.heCli, err = NewECDHE256(cfg.PriKey, "127.0.0.1", 1101, 1101)
	if err != nil {
		return nil, fmt.Errorf("create ecdhe client fail. %s", err.Error())
	}
	c.svr, err = ecies.NewENodeByUrl(cfg.EnodeUrl)
	if err != nil {
		return nil, fmt.Errorf("parse svr node fail. %s", err.Error())
	}
	c.notify = make(chan struct{})
	//1. 连接服务端
	go c.connect()
	go c.halfOpen()

	time.Sleep(time.Second)
	return c, nil
}

func (c *tcpc) connect() {
	for {
		//0. 初始化
		c.conn = nil
		//1. 连接失败 重试连接
		var address = fmt.Sprintf("%s:%d", c.svr.GetTcpHost(), c.svr.GetTcpPort())
		var conn, err = net.Dial("tcp", address)
		if err != nil {
			time.Sleep(50 * time.Millisecond)
			continue
		}
		//2. 连接成功, 协商密钥
		iRandPri, inonce, err := c.heCli.Write(conn, c.svr.GetPubKey())
		if err != nil {
			c.err = err
			return
		}
		svrPub, rRandPub, rnonce, err := c.heCli.Read(conn)
		if err != nil {
			c.err = err
			return
		}
		if !svrPub.Equal(c.svr.GetPubKey()) {
			c.err = fmt.Errorf("invalid node publick key")
			return
		}
		srt, err := c.heCli.Ephemeral(iRandPri, inonce, rRandPub, rnonce)
		if err != nil {
			c.err = err
			return
		}
		fmt.Printf("ak %x mk %x\n", srt.AK, srt.MK)
		sconn, err := NewConn(conn, &c.cfg.TransportOption, srt)
		if err != nil {
			time.Sleep(time.Second)
			continue
		}
		c.conn = sconn
		c.err = nil
		go c.heartbeat()
		select {
		case <-c.app.Done():
			conn.Close()
			c.conn = nil
			return
		case <-c.notify:
			conn.Close()
		}
		time.Sleep(time.Second)
	}
}

func (c *tcpc) heartbeat() {
	for {
		//1. 参数校验
		var sc = c.conn
		if sc == nil || c.err != nil {
			return
		}
		//2. 发Ping
		var _, err = sc.WriteRead(c.app, nil)
		if err == io.EOF {
			c.notify <- struct{}{}
			return
		}
		time.Sleep(time.Second / 5)
	}
}

func (c *tcpc) halfOpen() {
	var pp = make(chan os.Signal)
	signal.Notify(pp, syscall.SIGPIPE)
	select {
	case <-c.app.Done():
	case <-pp:
		c.notify <- struct{}{}
	}
}

func (c *tcpc) send(ctx context.Context, body []byte) error {
	//1. 获取加密通道
	var cc = c.conn
	if cc == nil || c.err != nil {
		return fmt.Errorf("conn fail. %v %v", cc, c.err)
	}
	//2. 压缩数据
	body, _ = c.cmpCli.Compress(body)
	//3. 传输数据
	return cc.Write(ctx, newmsg(body))
}

func (c *tcpc) receive(ctx context.Context) ([]byte, error) {
	//1. 获取加密通道
	var cc = c.conn
	if cc == nil || c.err != nil {
		return nil, fmt.Errorf("conn fail. %v %v", cc, c.err)
	}
	//2. 获取加密数据
	var m, err = cc.Read(ctx)
	if err != nil {
		return nil, err
	}
	//3. 解密数据
	m.data, _ = c.cmpCli.Uncompress(m.data)
	return m.data, nil
}

func (c *tcpc) Send(ctx context.Context, body []byte) error {
	select {
	case <-c.app.Done():
		return c.app.Err()
	default:
		return c.send(ctx, body)
	}
}

func (c *tcpc) Receive(ctx context.Context) ([]byte, error) {
	select {
	case <-c.app.Done():
		return nil, c.app.Err()
	default:
		return c.receive(ctx)
	}
}
