package dht

import (
	"hash/crc64"
	"net"
	"strconv"
)

type INode interface {
	Id() uint64
	Ipv4() uint32
	Port() uint16
	Zone() uint16
}

type enode struct {
	zone uint16
	port uint16
	ipv4 uint32
}

func NewNode(zone uint16, port uint16, ipv4 uint32) INode {
	var enr = &enode{}

	enr.ipv4 = ipv4
	enr.port = port
	enr.zone = zone

	return enr
}

func CRC(b []byte) uint64 {
	var table = crc64.MakeTable(crc64.ECMA)
	return crc64.Checksum(b, table)
}

func XOR(src INode, dst INode) uint8 {
	var (
		dist  = uint8(64)
		probe = uint64(0x8000000000000000)
		//距离的定义:
		// src: 101010101010110.........10101
		// dst: 101010101010111.........10101
		//  ^ : 000000000000001.........00000
		//		-----------------------------
		//	                  |--->dist<----|
		x = src.Id() ^ dst.Id()
	)

	for probe > 0 {
		if probe&x > 0 {
			break
		}
		probe >>= 1
		dist--
	}

	if dist >= 64 {
		dist = 63
	}

	if dist < 0 {
		dist = 0
	}

	return dist
}

func Encode(e INode) uint64 {
	return uint64(e.Zone())<<48 | uint64(e.Port())<<32 | uint64(e.Ipv4())
}

func Decode(a uint64) INode {
	return &enode{
		zone: uint16((a & 0xffff000000000000) >> 48),
		port: uint16((a & 0x0000ffff00000000) >> 32),
		ipv4: uint32(a & 0x00000000ffffffff),
	}
}

func (e *enode) Id() uint64 {
	var n = Encode(e)
	var b = make([]byte, 8)
	b[0] = byte(n >> 56)
	b[1] = byte(n >> 48)
	b[2] = byte(n >> 40)
	b[3] = byte(n >> 32)
	b[4] = byte(n >> 24)
	b[5] = byte(n >> 16)
	b[6] = byte(n >> 8)
	b[7] = byte(n)
	return CRC(b)
}

func (e *enode) Ipv4() uint32 {
	return e.ipv4
}

func (e *enode) Port() uint16 {
	return e.port
}

func (e *enode) Zone() uint16 {
	return e.zone
}

func NewNodeWithAddr(zone uint16, addr string) (INode, error) {
	var hostStr, portStr, err = net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	host := net.ParseIP(hostStr)
	if nil == host {
		return nil, errInvalidHost
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, errInvalidPort
	}
	host = host.To4()

	ipv4 := uint32(host[0]) << 24
	ipv4 |= uint32(host[1]) << 16
	ipv4 |= uint32(host[2]) << 8
	ipv4 |= uint32(host[3])

	return NewNode(zone, uint16(port), ipv4), nil
}
