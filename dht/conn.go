package dht

import "net"

type IDHTPacket interface {
	Addr() *net.UDPAddr
	Body() []byte
}

type packet struct {
	from *net.UDPAddr
	buf  []byte
}

func NewPacket(from *net.UDPAddr, body []byte) IDHTPacket {
	return &packet{from: from, buf: body}
}

func (p *packet) Addr() *net.UDPAddr {
	return p.from
}

func (p *packet) Body() []byte {
	return p.buf
}

type IDHTConn interface {
	ReadFromUDP(b []byte) (n int, addr *net.UDPAddr, err error)
	WriteToUDP(b []byte, addr *net.UDPAddr) (n int, err error)
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
