package dht

import (
	"context"
	"net"
	"strings"
)

type IDHTPacket interface {
	Ctx() context.Context
	Addr() *net.UDPAddr
	Body() []byte
}

type packet struct {
	ctx  context.Context
	addr *net.UDPAddr
	buf  []byte
}

func NewPacket(ctx context.Context, addr *net.UDPAddr, body []byte) IDHTPacket {
	return &packet{ctx: ctx, addr: addr, buf: body}
}

func (p *packet) Addr() *net.UDPAddr {
	return p.addr
}

func (p *packet) Body() []byte {
	return p.buf
}

func (p *packet) Ctx() context.Context {
	if nil == p.ctx {
		return context.TODO()
	} else {
		return p.ctx
	}
}

type IDHTConn interface {
	ReadFromUDP() ([]byte, *net.UDPAddr, error)
	WriteToUDP(b []byte, addr *net.UDPAddr) (n int, err error)
	Addr(node INode) *net.UDPAddr
	Close() error
}

type udpConn struct {
	cc    *net.UDPConn
	v     byte
	magic byte
}

func NewConn(network string, node INode) (IDHTConn, error) {
	if node == nil {
		return nil, errInvalidNode
	}
	switch strings.ToLower(network) {
	case "udp":
		return newUDPConn(node)
	default:
		return nil, errInvalidProtocol
	}
}

func newUDPConn(node INode) (IDHTConn, error) {
	var a = byte(node.Ipv4() >> 24)
	var b = byte(node.Ipv4() >> 16)
	var c = byte(node.Ipv4() >> 8)
	var d = byte(node.Ipv4())

	var ip = net.IPv4(a, b, c, d)
	var port = int(node.Port())

	cc, err := net.ListenUDP("udp", &net.UDPAddr{IP: ip, Port: port})
	if err != nil {
		return nil, err
	}
	return &udpConn{cc: cc, v: 0x01, magic: 0x71}, nil
}

//协议格式：Header + Body
//Header 4Byte
// 0 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20 21 22 23 24 25 26 27 28 29 30 31
// |-------------|---------------------|------------------------------------------------|
//     magic	   version |  padding			 length
func (c *udpConn) ReadFromUDP() ([]byte, *net.UDPAddr, error) {
	var b = make([]byte, maxPacketSize)

	var n, addr, err = c.cc.ReadFromUDP(b)
	if err != nil {
		return nil, nil, err
	}
	b = b[:n]

	var (
		magic   = b[0]
		ver     = (b[1] & 0xf0) >> 4
		padding = b[1] & 0x0f
	)
	if magic != c.magic || ver != c.v {
		return nil, nil, errInvalidMessage
	}
	var pl = int(b[2]<<8) | int(b[3])
	b = b[4 : pl-int(padding)]

	return b, addr, nil
}

func (c *udpConn) WriteToUDP(b []byte, addr *net.UDPAddr) (n int, err error) {
	var (
		bl      = len(b)
		padding = bl % 16
	)

	if padding > 0 {
		padding = 16 - padding
	}
	var pl = bl + padding + 0x04
	var p = make([]byte, pl)

	p[0] = c.magic
	p[1] = byte(c.v<<4)&0xf0 | byte(padding)&0x0f
	p[2] = byte(pl >> 8)
	p[3] = byte(pl)
	copy(p[0x04:], b)

	return c.cc.WriteToUDP(p, addr)
}

func (c *udpConn) Close() error {
	return c.cc.Close()
}

func (cc udpConn) Addr(node INode) *net.UDPAddr {
	var a = byte(node.Ipv4() >> 24)
	var b = byte(node.Ipv4() >> 16)
	var c = byte(node.Ipv4() >> 8)
	var d = byte(node.Ipv4())

	var ip = net.IPv4(a, b, c, d)
	var port = int(node.Port())

	return &net.UDPAddr{IP: ip, Port: port}
}
