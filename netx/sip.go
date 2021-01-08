package netx

import (
	"context"
	"sync"
	"time"

	"github.com/advancevillage/3rd/utils"
)

//sip is session identifier protocol
type ISIPClient interface {
	Send(context.Context, []byte) ([]byte, error)
}

type pkt []byte

func (p *pkt) Bytes() []byte {
	return *p
}

type sipClient struct {
	tcpCli ITcpClient
	mu     sync.RWMutex
	sip    map[string]chan pkt
	cfg    *TcpClientOpt
	app    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	sLen   int //session id len uuid 36
}

func NewSIPClent(cfg *TcpClientOpt) (ISIPClient, error) {
	var c = &sipClient{}
	var err error
	c.tcpCli, err = NewTcpClient(cfg)
	if err != nil {
		return nil, err
	}
	c.sip = make(map[string]chan pkt)
	c.cfg = cfg
	c.app, c.cancel = context.WithCancel(context.TODO())
	c.sLen = 36

	go c.readLoop()

	return c, nil
}

func (c *sipClient) Send(ctx context.Context, body []byte) ([]byte, error) {
	return c.write(ctx, body)
}

func (c *sipClient) readLoop() {
	for {
		select {
		case <-c.app.Done():
			return
		default:
			var b, err = c.tcpCli.Receive(c.app)
			if err != nil {
				continue
			}
			go c.read(b)
		}
	}
}

func (c *sipClient) read(b []byte) {
	c.wg.Add(1)
	defer c.wg.Done()
	switch {
	case len(b) < c.sLen:
	default:
		var sid = string(b[:c.sLen])
		var p = pkt(b[c.sLen:])
		var ch, ok = c.get(sid)
		if ok {
			ch <- p
		}
		//session timeout drop pkt
	}
}

func (c *sipClient) write(ctx context.Context, b []byte) ([]byte, error) {
	var (
		t    = time.NewTicker(c.cfg.Timeout)
		sid  = utils.UuId()
		body = make([]byte, len(sid)+len(b))
		err  error
	)
	defer t.Stop()
	copy(body[:len(sid)], sid)
	copy(body[len(sid):], b)
	var ch = c.set(string(sid))
	err = c.tcpCli.Send(ctx, body)
	if err != nil {
		return nil, err
	}
	//注册sid
	select {
	case <-t.C: //session id timeout
		c.del(string(sid))
		return nil, errSessionTimeout
	case <-ctx.Done():
		c.del(string(sid))
		return nil, ctx.Err()
	case <-c.app.Done():
		c.del(string(sid))
		return nil, c.app.Err()
	case pkt := <-ch:
		return pkt.Bytes(), nil
	}
}

func (c *sipClient) set(sid string) chan pkt {
	var ch = make(chan pkt)
	c.mu.Lock()
	c.sip[sid] = ch
	c.mu.Unlock()
	return ch
}

func (c *sipClient) del(sid string) {
	c.mu.Lock()
	delete(c.sip, sid)
	c.mu.Unlock()
}

func (c *sipClient) get(sid string) (ch chan pkt, ok bool) {
	c.mu.RLock()
	ch, ok = c.sip[sid]
	c.mu.RUnlock()
	return
}

func (c *sipClient) close() {
	c.cancel()
}

//session tcp server
type ISIPServer interface {
	ITcpServer
}

type sipServer struct {
	tcpSvr ITcpServer
	cfg    *TcpServerOpt
	sLen   int
}

func NewSIPServer(cfg *TcpServerOpt) (ISIPServer, error) {
	var s = &sipServer{}
	var err error
	var tCfg = &TcpServerOpt{
		Host: cfg.Host,
		Port: cfg.Port,
		PC:   cfg.PC,
		PH:   s.handle,
	}

	s.cfg = cfg
	s.sLen = 36
	s.tcpSvr, err = NewTcpServer(tCfg)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (s *sipServer) StartServer() {
	s.tcpSvr.StartServer()
}

func (s *sipServer) StopServer() {
	s.tcpSvr.StopServer()
}

func (s *sipServer) handle(ctx context.Context, body []byte) []byte {
	switch {
	case len(body) < s.sLen: //drop pkt
		return nil
	default:
		var sid = body[:s.sLen]
		var b = body[s.sLen:]
		b = s.cfg.PH(ctx, b)
		var p = make([]byte, s.sLen+len(b))
		copy(p[:s.sLen], sid)
		copy(p[s.sLen:], b)
		return p
	}
}
