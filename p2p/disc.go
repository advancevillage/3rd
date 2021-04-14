package p2p

import (
	"context"
	"fmt"
	"time"

	"github.com/advancevillage/3rd/ecies"
	"github.com/advancevillage/3rd/netx"
)

const (
	TypePing      = byte(0x02)
	TypeFindNode  = byte(0x03)
	TypeFindValue = byte(0x04)
	TypeError     = byte(0xff)
)

type IKad interface {
	Start()
}

type kad struct {
	udpq    IP2PQueue
	dhtq    IP2PQueue
	ctx     context.Context
	udp     netx.IUDPServer
	timeout time.Duration
}

func NewKad(ctx context.Context, local ecies.IENode, udpq IP2PQueue, dhtq IP2PQueue) (IKad, error) {
	//1. 参数校验
	var (
		err error
		ks  = &kad{}
		cfg = &netx.ServerOption{}
	)
	cfg.Host = local.GetTcpHost()
	cfg.UdpPort = local.GetUdpPort()
	cfg.MaxSize = uint32(176 * 25)
	ks.udp, err = netx.NewUDPServer(cfg, ks.handler)
	ks.udpq = udpq
	ks.dhtq = dhtq
	ks.ctx = ctx
	ks.timeout = time.Second * 3
	if err != nil {
		return nil, err
	}
	if ks.dhtq == nil || ks.udpq == nil {
		return nil, fmt.Errorf("p2p queue is nil")
	}
	return ks, nil
}

//@overview: kad 协议处理器
//@author: richard.sun
//@param: pro 服务端收到的协议内容
//   [0][0 0][........]
//   type len  context
//type: 协议类型 1个字节
//len:  内容长度 2个字节 最大长度65535
//context: 协议内容
func (ks *kad) handler(ctx context.Context, ss []byte) []byte {
	var tYpe byte
	switch {
	case len(ss) <= 0:
		return ks.write(TypeError, []byte("protocol type is empty"))
	case ss[0] == TypePing:
		tYpe = TypePing
	case ss[0] == TypeFindNode:
		tYpe = TypeFindNode
	case ss[0] == TypeFindValue:
		tYpe = TypeFindValue
	default:
		return ks.write(TypeError, []byte("don't support protocol type"))
	}
	var body, err = ks.parse(ss)
	if err != nil {
		return ks.write(TypeError, []byte("protocol type is empty"))
	}
	body, err = ks.push(ctx, tYpe, body)
	if err != nil {
		return ks.write(TypeError, []byte("protocol type is empty"))
	}
	return body
}

func (ks *kad) write(tYpe byte, d []byte) []byte {
	var le = len(d)
	var e = make([]byte, 0x3+le)
	var l = make([]byte, 2)
	ks.writeInt16(uint16(le), l)
	e[0] = tYpe
	copy(e[0x1:], l)
	copy(e[0x3:], d)
	return e
}

func (ks *kad) parse(in []byte) ([]byte, error) {
	//1. 解析参数
	if len(in) < 0x3 {
		return nil, fmt.Errorf("protocol len is empty")
	}
	var le = ks.readInt16(in[0x1:0x3])
	if 0x3+le > uint16(len(in)) {
		return nil, fmt.Errorf("protocol context is incomplete")
	}
	//2. 解析内容
	return in[0x3 : 0x3+le], nil
}

func (ks *kad) push(ctx context.Context, tYpe byte, body []byte) ([]byte, error) {
	var m = newmsg(tYpe, body, nil)
	var t = time.NewTicker(ks.timeout)
	ks.dhtq.Push(m)
	select {
	case body = <-m.out:
		return body, nil
	case <-t.C:
		return nil, fmt.Errorf("data handle timeout")
	}
}

func (ks *kad) readInt16(b []byte) uint16 {
	return uint16(b[1]) | uint16(b[0])<<8
}

func (ks *kad) writeInt16(v uint16, b []byte) {
	b[0] = byte(v >> 8)
	b[1] = byte(v)
}

//todo: 缓存连接
func (ks *kad) dial(m *msg) {
	var cfg = &netx.ClientOption{}
	cfg.Host = m.enode.GetTcpHost()
	cfg.UdpPort = m.enode.GetUdpPort()
	cfg.MaxSize = uint32(176 * 25)
	c, err := netx.NewUDPClient(cfg)
	if err != nil {
		return
	}
	p, err := c.SendReceive(ks.ctx, ks.write(m.tYpe, m.in))
	if err != nil {
		return
	}
	ss, err := ks.parse(p)
	if err != nil {
		return
	}
	m.out <- ss
}

func (ks *kad) loop() {
	for {
		m := <-ks.udpq.Pull()
		if m == nil || m.enode == nil || m.out == nil {
			continue
		}
		go ks.dial(m)
	}
}

func (ks *kad) Start() {
	go ks.loop()
	ks.udp.StartServer()
}
