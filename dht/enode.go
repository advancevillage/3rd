package dht

import (
	"strings"
)

type INode interface {
	Encode() uint64
	Decode(a uint64) INode

	Ipv4() uint32
	Port() uint16
	Zone() uint16
	Protocol() string
}

type enode struct {
	zone uint16 //12bit zone; 4bit protocol
	port uint16
	ipv4 uint32
}

func NewNode(protocol string, zone uint16, port uint16, ipv4 uint32) INode {
	var enr = &enode{}

	enr.ipv4 = ipv4
	enr.port = port
	enr.zone = zone & 0xfff0

	switch strings.ToLower(protocol) {
	case "tcp":
		enr.zone |= 0x0001
	case "udp":
		enr.zone |= 0x0002
	}
	return enr
}

func (e *enode) Encode() uint64 {
	return uint64(e.zone)<<48 | uint64(e.port)<<32 | uint64(e.ipv4)
}

func (e enode) Decode(a uint64) INode {
	return &enode{
		zone: uint16((a & 0xffff000000000000) >> 48),
		port: uint16((a & 0x0000ffff00000000) >> 32),
		ipv4: uint32(a & 0x00000000ffffffff),
	}
}

func (e *enode) Ipv4() uint32 {
	return e.ipv4
}

func (e *enode) Port() uint16 {
	return e.port
}

func (e *enode) Zone() uint16 {
	return e.zone & 0xfff0
}

func (e *enode) Protocol() string {
	var p string
	switch e.zone & 0x000f {
	case 0x0001:
		p = "tcp"
	case 0x0002:
		p = "udp"
	}
	return p
}
