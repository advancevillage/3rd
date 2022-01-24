package dht

import (
	"context"
	"net"
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
	ReadFromUDP(b []byte) (n int, addr *net.UDPAddr, err error)
	WriteToUDP(b []byte, addr *net.UDPAddr) (n int, err error)
	Addr(node INode) *net.UDPAddr
	Close() error
}

type udpConn struct {
	*net.UDPConn
}

func NewUDPConn(node INode) (IDHTConn, error) {
	if node == nil || node.Protocol() != "udp" {
		return nil, errInvalidNode
	}

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
	return &udpConn{cc}, nil
}

func (c *udpConn) ReadFromUDP(b []byte) (n int, addr *net.UDPAddr, err error) {
	return c.ReadFromUDP(b)
}

func (c *udpConn) WriteToUDP(b []byte, addr *net.UDPAddr) (n int, err error) {
	return c.WriteToUDP(b, addr)
}

func (c *udpConn) Close() error {
	return c.Close()
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
