package netx

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/advancevillage/3rd/utils"
)

var (
	zero4 = []byte{0x0, 0x0, 0x0, 0x0}
)

//@overview: 传输层接口协议. 特点是安全,灵活,高效
//@author: richard.sun
//@note:
//1. write frame
//2. read frame
type ITransport interface {
	Read(context.Context) ([]byte, error)
	Write(context.Context, []byte) error
	WriteRead(context.Context, []byte) ([]byte, error)
}

type TransportOption struct {
	Host    string
	Port    int
	UdpPort int
	MaxSize uint32
	Timeout time.Duration     //读写超时
	PriKey  *ecdsa.PrivateKey //本地服务私钥
}

//@overview: 并发安全
type tcpConn struct {
	imac ITcpMac          //加密通讯通道
	conn net.Conn         //TCP连接
	cfg  *TransportOption //配置文件

	hSize  int //协议头数据长度
	pad    int //对齐长度
	rc     chan *msg
	wc     chan *msg
	sip    map[string]chan *msg
	rwm    sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
}

type msg struct {
	fId   []byte
	flags []byte
	data  []byte
	err   error
}

func NewConn(conn net.Conn, cfg *TransportOption, esrt *Secrets) (ITransport, error) {
	//1. 参数校验
	if nil == conn {
		return nil, fmt.Errorf("tcp conn is invalid")
	}
	if nil == esrt {
		return nil, fmt.Errorf("tcp ephemeral secret is invalid")
	}
	if nil == cfg {
		return nil, fmt.Errorf("tcp cfg is invalid")
	}
	if cfg.MaxSize <= 0 {
		return nil, fmt.Errorf("tcp cfg maxSize is invalid")
	}
	if nil == cfg.PriKey {
		return nil, fmt.Errorf("tcp cfg private key is invalid")
	}
	//2. 构建加密通道
	var mac, err = NewTcpMac(*esrt)
	if err != nil {
		return nil, fmt.Errorf("tcp ephemeral secret mac fail. %s", err.Error())
	}
	var ctx, cancel = context.WithCancel(context.TODO())
	var cc = &tcpConn{
		imac:   mac,
		conn:   conn,
		cfg:    cfg,
		hSize:  0x20,
		pad:    0x10,
		ctx:    ctx,
		cancel: cancel,
		rc:     make(chan *msg, 1024),
		wc:     make(chan *msg, 1024),
		sip:    make(map[string]chan *msg),
	}

	go cc.readLoop()
	go cc.writeLoop()

	return cc, nil
}

func (c *tcpConn) readLoop() {
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			//1.读取数据
			var flags, fId, data, err = c.imac.ReadFrame(c.conn, c.hSize, c.pad)
			var m = &msg{
				flags: flags,
				fId:   fId,
				err:   err,
				data:  data,
			}
			var t = time.NewTicker(time.Second)
			if ch, ok := c.get(m.fId); ok {
				select {
				case ch <- m:
				case <-t.C:
				}
			} else {
				select {
				case c.rc <- m:
				case <-t.C:
				}
			}
		}
	}
}

func (c *tcpConn) writeLoop() {
	for m := range c.wc {
		select {
		case <-c.ctx.Done():
			return
		default:
			c.imac.WriteFrame(c.conn, c.hSize, c.pad, m.flags, m.fId, m.data)
		}
	}
}

func (c *tcpConn) set(fId []byte) chan *msg {
	var ch = make(chan *msg)
	c.rwm.Lock()
	c.sip[string(fId)] = ch
	c.rwm.Unlock()
	return ch
}

func (c *tcpConn) del(fId []byte) {
	c.rwm.Lock()
	delete(c.sip, string(fId))
	c.rwm.Unlock()
}

func (c *tcpConn) get(fId []byte) (chan *msg, bool) {
	c.rwm.RLock()
	ch, ok := c.sip[string(fId)]
	c.rwm.RUnlock()
	return ch, ok
}

func (c *tcpConn) Read(ctx context.Context) ([]byte, error) {
	var t = time.NewTicker(c.cfg.Timeout)
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-c.ctx.Done():
		return nil, ctx.Err()
	case <-t.C:
		return nil, fmt.Errorf("read data timeout. consume more than %d's", c.cfg.Timeout/time.Second)
	case m := <-c.rc: //1. 读取数据
		return m.data, m.err
	}
}

func (c *tcpConn) Write(ctx context.Context, data []byte) error {
	//1. 校验
	if len(data) > int(c.cfg.MaxSize) {
		return fmt.Errorf("package size more than %d", c.cfg.MaxSize)
	}
	//2.
	var m = &msg{
		data:  data,
		fId:   utils.UUID8Byte(),
		flags: zero4,
	}
	var t = time.NewTicker(c.cfg.Timeout)
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-c.ctx.Done():
		return c.ctx.Err()
	case <-t.C:
		return fmt.Errorf("write data timeout. consume more than %d's", c.cfg.Timeout/time.Second)
	case c.wc <- m:
		return nil
	}
}

func (c *tcpConn) WriteRead(ctx context.Context, data []byte) ([]byte, error) {
	//1. 校验
	if len(data) > int(c.cfg.MaxSize) {
		return nil, fmt.Errorf("package size more than %d", c.cfg.MaxSize)
	}
	//2.
	var m = &msg{
		data:  data,
		fId:   utils.UUID8Byte(),
		flags: zero4,
	}
	//3. write data
	var (
		t  = time.NewTicker(c.cfg.Timeout)
		ch chan *msg
	)

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-c.ctx.Done():
		return nil, c.ctx.Err()
	case <-t.C:
		return nil, fmt.Errorf("write data timeout. more than %d's", c.cfg.Timeout/time.Second)
	case c.wc <- m:
		ch = c.set(m.fId)
	}
	defer c.del(m.fId)
	//4. read data
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-c.ctx.Done():
		return nil, c.ctx.Err()
	case <-t.C:
		return nil, fmt.Errorf("read data timeout. more than %d's", c.cfg.Timeout/time.Second)
	case v := <-ch:
		return v.data, nil
	}
}
