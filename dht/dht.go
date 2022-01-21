package dht

import (
	"context"
	"errors"
	"io"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/advancevillage/3rd/logx"
	"github.com/advancevillage/3rd/proto"
	"github.com/advancevillage/3rd/radix"
	enc "github.com/golang/protobuf/proto"
)

var (
	errInvalidHost    = errors.New("invalid host")
	errInvalidPort    = errors.New("invalid port")
	errInvalidNode    = errors.New("invalid node")
	errInvalidMessage = errors.New("invalid message")

	maxPacketSize = 1024
)

type IDHT interface {
}

func NewDHT(ctx context.Context, logger logx.ILogger, zone uint16, network string, addr string) (IDHT, error) {
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

	ipv4 := uint32(host[0]) << 24
	ipv4 |= uint32(host[1]) << 16
	ipv4 |= uint32(host[2]) << 8
	ipv4 |= uint32(host[3])

	node := NewNode(network, zone, uint16(port), ipv4)

	conn, err := NewUDPConn(node)
	if err != nil {
		return nil, err
	}

	var d = &dht{}
	d.self = node
	d.ctx, d.canecl = context.WithCancel(ctx)
	d.logger = logger

	d.k = 4
	d.bit = 0x40
	d.alpha = 3
	d.bucket = make([]uint8, d.bit)

	d.rtm = radix.NewRadixTree()
	d.nodes = make(map[uint64]*dhtNode)

	d.refresh = 10
	d.quit = make(chan bool, 1)

	d.conn = conn
	d.packetInCh = make(chan IDHTPacket, 1)
	d.readNextCh = make(chan struct{}, 1)

	return d, nil
}

type dht struct {
	self   INode
	ctx    context.Context
	canecl context.CancelFunc
	logger logx.ILogger

	//常数设置
	k      uint8
	alpha  uint8
	bit    uint8
	bucket []uint8

	//路由表
	rtm   radix.IRadixTree
	nodes map[uint64]*dhtNode //存储节点列表
	rwm   sync.RWMutex

	//事件指令
	refresh int
	quit    chan bool

	//通讯协议
	conn       IDHTConn
	packetInCh chan IDHTPacket
	readNextCh chan struct{}
}

type dhtNode struct {
	node  INode     //节点
	at    time.Time //节点加入时间
	delay uint64    //毫秒
	succ  uint8     //PING回复次数
	total uint8     //PING发送次数
}

func (d *dht) loop() {
	var (
		refreshC = time.NewTicker(time.Second * time.Duration(d.refresh))
		fixC     = time.NewTicker(time.Second * time.Duration(d.refresh*60))
	)
	defer refreshC.Stop()
	defer fixC.Stop()

	for {
		select {
		case <-d.ctx.Done(): //退出事件
			goto end

		case <-d.quit:
			goto end

		case <-refreshC.C: //刷新事件
			d.doRefresh()

		case <-fixC.C: //修复事件
			d.doFix()

		case p := <-d.packetInCh:
			d.doPacket(p)
			d.readNextCh <- struct{}{}

		}
	}

end:
	d.doGc()
}

func (d *dht) doPacket(pkt IDHTPacket) {
	var (
		msg = new(proto.Packet)
		err = enc.Unmarshal(pkt.Body(), msg)
	)
	if err != nil {
		d.logger.Warnw(d.ctx, errInvalidMessage.Error(), "err", err)
		return
	}

	var (
		ctx = context.WithValue(d.ctx, logx.TraceId, msg.GetTrace())
	)
	switch msg.GetType() {
	case proto.PacketType_null:

	case proto.PacketType_ping:
		var ping = new(proto.Ping)
		err = enc.Unmarshal([]byte(msg.GetPkt()), ping)
		if err != nil {
			d.logger.Warnw(ctx, errInvalidMessage.Error(), "err", err)
		}
		d.doPing(ctx, ping)
	case proto.PacketType_store:

	case proto.PacketType_findnode:

	case proto.PacketType_findvalue:

	}
}

func (d *dht) doPing(ctx context.Context, msg *proto.Ping) {
	//1. 检查目的节点是不是给自己的
	if msg.GetTo() != d.self.Encode() {
		return
	}
	//2. 检查消息状态
	switch msg.GetState() {
	case proto.State_ack:

	case proto.State_syn:
		msg.State = proto.State_ack
		msg.Rt = uint64(time.Now().UnixNano())
		msg.To = msg.From
		msg.From = d.self.Encode()
		//TODO: send

	}

}

func (d *dht) doRefresh() {

}

func (d *dht) doFix() {

}

func (d *dht) doGc() {
	//1. 回收通讯FD
	if nil != d.conn {
		d.conn.Close()
	}
}

func (d *dht) xor(n INode) uint8 {
	var (
		dist = d.bit
		bit  = uint64(0x8000000000000000)
		x    = d.self.Encode() ^ n.Encode()
	)

	for bit > 0 && x > 0 {
		if bit&x > 0 {
			break
		}
		bit >>= 1
		dist--
	}

	return d.bit - dist
}

func (s *dht) readLoop() {
	var buf = make([]byte, maxPacketSize)
	for range s.readNextCh {
		n, from, err := s.conn.ReadFromUDP(buf)
		if err == io.EOF {
			s.logger.Errorw(s.ctx, "dht srv read eof")
			//TODO 推出或者重启
			return
		}
		if nerr, ok := err.(net.Error); ok && nerr.Temporary() {
			s.logger.Warnw(s.ctx, "dht srv read temporary")
			time.Sleep(time.Second / 10)
			continue
		}
		s.dispatchReadPacket(from, buf[:n])
	}
}

func (d *dht) dispatchReadPacket(from *net.UDPAddr, body []byte) bool {
	select {
	case d.packetInCh <- NewPacket(from, body):
		return true
	case <-d.ctx.Done():
		return false
	}
}
