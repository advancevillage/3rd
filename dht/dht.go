package dht

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/advancevillage/3rd/logx"
	"github.com/advancevillage/3rd/mathx"
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

type kpiL uint8

var (
	kpiSSS = kpiL(0x01) // <= 5ms & fail <= 2
	kpiSS  = kpiL(0x02) // <= 10ms & fail <= 2
	kpiS   = kpiL(0x03) // <=  20ms && fail <= 4
	kpiA   = kpiL(0x04) // <= 40ms && fail <= 4
	kpiB   = kpiL(0x05) // <= 80ms && fail <= 4
	kpiC   = kpiL(0x06) // <= 160ms && fail <= 4
	kpiD   = kpiL(0x07) // <=320ms && fail <= 4
	kpiE   = kpiL(0x08) //<=640ms && fail <= 6
	kpiF   = kpiL(0x09) // <= 1000ms && fail <= 6
	kpiG   = kpiL(0x0a) // <= 1500ms && fail <= 6
	kpiH   = kpiL(0x0b) // <= 2000ms && fail <= 6
	kpiI   = kpiL(0x0c) // <= 2500ms && fail <= 6
	kpiJ   = kpiL(0x0e) // <= 3000ms && fail <= 6
	kpiK   = kpiL(0x0f) // <= 3500ms && fail <= 6
	kpiM   = kpiL(0x10) // <= 4000ms && fail <= 6
	kpiZ   = kpiL(0xff) // <= 4000ms && fail <= 6
)

type IDHT interface {
	Start()
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
	host = host.To4()

	ipv4 := uint32(host[0]) << 24
	ipv4 |= uint32(host[1]) << 16
	ipv4 |= uint32(host[2]) << 8
	ipv4 |= uint32(host[3])

	fmt.Println(ipv4)

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

	d.conn = conn
	d.packetInCh = make(chan IDHTPacket, d.alpha)
	d.packetOutCh = make(chan IDHTPacket, d.alpha)

	go d.readLoop()
	go d.loop()
	go d.store(d.self)

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

	//通讯协议
	conn        IDHTConn
	packetInCh  chan IDHTPacket
	packetOutCh chan IDHTPacket
}

type dhtNode struct {
	node  INode     //节点
	at    time.Time //节点加入时间
	delay int64     //纳秒
	succ  int64     //接受次数
	total int64     //发送次数
	keep  int64     //连续保持状态次数
	score kpiL
}

func newDHTNode(node INode, initScore kpiL) *dhtNode {
	return &dhtNode{
		node:  node,
		at:    time.Now(),
		total: 0,
		succ:  0,
		delay: 0,
		keep:  0,
		score: initScore,
	}
}

func (n *dhtNode) kpi() {
	var delay = n.delay / 1e6
	var fail = n.total - n.succ
	var last = n.score

	switch {
	case delay <= 5 && fail <= 2:
		n.score = kpiSSS
	case delay <= 10 && fail <= 2:
		n.score = kpiSS
	case delay <= 20 && fail <= 4:
		n.score = kpiS
	case delay <= 40 && fail <= 4:
		n.score = kpiA
	case delay <= 80 && fail <= 4:
		n.score = kpiB
	case delay <= 160 && fail <= 4:
		n.score = kpiC
	case delay <= 320 && fail <= 4:
		n.score = kpiD
	case delay <= 640 && fail <= 6:
		n.score = kpiE
	case delay <= 1000 && fail <= 6:
		n.score = kpiF
	case delay <= 1500 && fail <= 6:
		n.score = kpiG
	case delay <= 2000 && fail <= 6:
		n.score = kpiH
	case delay <= 2500 && fail <= 6:
		n.score = kpiI
	case delay <= 3000 && fail <= 6:
		n.score = kpiJ
	case delay <= 3500 && fail <= 6:
		n.score = kpiK
	case delay <= 4000 && fail <= 6:
		n.score = kpiM
	default:
		n.score = kpiZ
	}

	if last == n.score {
		n.keep++
	} else {
		n.keep = 0
	}

	switch {
	case n.keep >= 0x20 && fail > 0:
		n.keep -= 0x20
		n.succ++
	}
}

func (d *dht) Start() {
	d.logger.Infow(d.ctx, "dht srv start", "node", fmt.Sprintf("%x", d.self.Encode()), "addr", d.conn.Addr(d.self))
	select {
	case <-d.ctx.Done():
		d.logger.Infow(d.ctx, "dht srv exit")
	}
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
			goto loopEnd

		case <-refreshC.C: //刷新事件
			d.doRefresh()

		case <-fixC.C: //修复事件
			d.doFix()

		case p := <-d.packetInCh:
			d.doReceive(p)

		case p := <-d.packetOutCh:
			d.doSend(p)
		}
	}

loopEnd:
	d.logger.Infow(d.ctx, "dht srv main loop end")

	d.doGc()
}

func (d *dht) doSend(pkt IDHTPacket) {
	var n = len(pkt.Body())

	if n > maxPacketSize {
		d.logger.Warnw(pkt.Ctx(), "dht srv message exceeds the maximum value", "cur", n, "max", maxPacketSize)
		return
	}

	nn, err := d.conn.WriteToUDP(pkt.Body(), pkt.Addr())
	if err != nil {
		d.logger.Warnw(pkt.Ctx(), "dht srv send message fail", "err", err)
		return
	}

	if nn < n {
		d.logger.Warnw(pkt.Ctx(), "dht srv message sent incomplete", "sent", nn, "should", n)
	}
}

func (d *dht) doReceive(pkt IDHTPacket) {
	var (
		msg = new(proto.Packet)
		err = enc.Unmarshal(pkt.Body(), msg)
	)
	if err != nil {
		d.logger.Warnw(d.ctx, errInvalidMessage.Error(), "err", err)
		return
	}

	var (
		ctx = context.WithValue(pkt.Ctx(), logx.TraceId, msg.GetTrace())
	)
	switch msg.GetType() {
	case proto.PacketType_null:

	case proto.PacketType_ping:
		var ping = new(proto.Ping)
		err = enc.Unmarshal([]byte(msg.GetPkt()), ping)
		if err != nil {
			d.logger.Warnw(ctx, errInvalidMessage.Error(), "err", err)
			return
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
	var now = time.Now().UnixNano()
	//2. 检查消息状态
	switch msg.GetState() {
	case proto.State_ack:
		value, ok := d.nodes[msg.From]
		if !ok {
			return
		}
		value.succ++
		value.delay = now - msg.St
		value.kpi()

	case proto.State_syn:
		msg.State = proto.State_ack
		msg.Rt = now
		msg.To = msg.From
		msg.From = d.self.Encode()

		buf, err := enc.Marshal(msg)
		if err != nil {
			d.logger.Warnw(ctx, "dht srv marshal ping message", "err", err)
			return
		}

		pkt := &proto.Packet{
			Type: proto.PacketType_ping,
			Pkt:  string(buf),
		}

		d.send(ctx, d.self.Decode(msg.To), pkt)
	}
}

func (d *dht) doRefresh() {
	d.rwm.RLock()
	defer d.rwm.RUnlock()
	var now = time.Now().UnixNano()
	var ctx = context.WithValue(d.ctx, logx.TraceId, mathx.UUID())

	for node, value := range d.nodes {
		var msg = new(proto.Ping)

		msg.From = d.self.Encode()
		msg.To = node
		msg.St = now
		msg.State = proto.State_syn

		buf, err := enc.Marshal(msg)
		if err != nil {
			d.logger.Warnw(ctx, "dht srv marshal ping message", "err", err, "node", node)
			return
		}

		pkt := &proto.Packet{
			Type: proto.PacketType_ping,
			Pkt:  string(buf),
		}

		d.send(ctx, d.self.Decode(node), pkt)
		value.total++
	}
}

func (d *dht) doFix() {

}

func (d *dht) doGc() {
	//1. 回收通讯FD
	if nil != d.conn {
		d.conn.Close()
	}
}

func (d *dht) readLoop() {
	var buf = make([]byte, maxPacketSize)

	for {
		n, from, err := d.conn.ReadFromUDP(buf)

		if err == io.EOF {
			d.logger.Errorw(d.ctx, "dht srv read eof")
			goto readLoopEnd
		}

		if nerr, ok := err.(net.Error); ok && nerr.Temporary() {
			d.logger.Warnw(d.ctx, "dht srv read temporary")
			time.Sleep(time.Second / 10)
			continue
		}

		select {
		case d.packetInCh <- NewPacket(d.ctx, from, buf[:n]):
		case <-d.ctx.Done():
			goto readLoopEnd
		}

	}

readLoopEnd:
	d.logger.Infow(d.ctx, "dnt srv readLoop end")
}

func (d *dht) send(ctx context.Context, to INode, pkt *proto.Packet) {

	if to.Encode() == d.self.Encode() {
		return
	}

	pkt.Trace = mathx.UUID()
	buf, err := enc.Marshal(pkt)
	if err != nil {
		d.logger.Warnw(ctx, errInvalidMessage.Error(), "err", err)
		return
	}

	select {
	case d.packetOutCh <- NewPacket(ctx, d.conn.Addr(to), buf):
	case <-d.ctx.Done():
		return
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

func (d *dht) store(node INode) {
	if node == nil {
		return
	}
	var err = d.rtm.AddU64(node.Encode(), 0xffffffffffffffff, node.Encode())
	if err != nil {
		d.logger.Warnw(d.ctx, "dht srv store node fail", "err", err, "node", node.Encode())
		return
	}
	var initScore kpiL
	if node.Encode() == d.self.Encode() {
		initScore = kpiSSS
	} else {
		initScore = kpiZ
	}

	d.rwm.Lock()
	d.nodes[node.Encode()] = newDHTNode(node, initScore)
	d.rwm.Unlock()
}