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
	Monitor() interface{}
	Start()
}

func (d *dht) Start() {
	d.logger.Infow(d.ctx, "dht srv start", "node", fmt.Sprintf("%x", d.self.Encode()), "addr", d.conn.Addr(d.self))

	go d.readLoop()
	go d.loop()
	go d.store(d.ctx, d.self)
	go d.seeds(d.ctx, d.seed)

	select {
	case <-d.ctx.Done():
		d.logger.Infow(d.ctx, "dht srv exit")
	}
}

func (d *dht) Monitor() interface{} {
	return d.dump(d.ctx)
}

func NewDHT(ctx context.Context, logger logx.ILogger, zone uint16, network string, addr string, seed []uint64) (IDHT, error) {
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
	d.factor = 10
	d.bucket = make([]uint8, d.bit)
	d.seed = seed

	d.rtm = radix.NewRadixTree()
	d.nodes = make(map[uint64]*dhtNode)

	d.conn = conn
	d.packetInCh = make(chan IDHTPacket, d.alpha)
	d.packetOutCh = make(chan IDHTPacket, d.alpha)

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
	factor int

	//路由表
	rtm   radix.IRadixTree
	nodes map[uint64]*dhtNode //存储节点列表
	rwm   sync.RWMutex
	wg    sync.WaitGroup
	seed  []uint64 //初始节点

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

	if last == n.score && n.score != kpiZ {
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

func (d *dht) loop() {
	go d.loopTimer()
	go d.loopNetIO()
	go d.loopRPC()
	d.wg.Wait()

	//回收事件
	d.doGc(d.ctx)
}

func (d *dht) loopNetIO() {
	d.wg.Add(1)
	defer d.wg.Done()

	for {
		select {
		//收报事件
		case p := <-d.packetInCh:
			d.doReceive(p.Ctx(), p)

		//发报事件
		case p := <-d.packetOutCh:
			d.doSend(p.Ctx(), p)

		case <-d.ctx.Done():
			goto loopIOEnd
		}
	}

loopIOEnd:
	d.logger.Infow(d.ctx, "dht srv main loop io end")

}

func (d *dht) loopTimer() {
	d.wg.Add(1)
	defer d.wg.Done()

	var (
		t       = mathx.Primes(d.factor)
		lt      = len(t) - 1
		fixC    = time.NewTicker(time.Second * time.Duration(t[lt-3]))
		evolutC = time.NewTimer(time.Second * time.Duration(t[lt]))
	)
	defer fixC.Stop()
	defer evolutC.Stop()

	for {
		select {
		//退出事件
		case <-d.ctx.Done():
			goto loopTimerEnd

		//修复事件
		case <-fixC.C:
			d.doFix(d.ctx)

		//进化事件
		case <-evolutC.C:
			d.doEvolut(d.ctx)
		}
	}

loopTimerEnd:
	d.logger.Infow(d.ctx, "dht srv main loop timer end")
}

func (d *dht) loopRPC() {
	d.wg.Add(1)
	defer d.wg.Done()

	var (
		t        = mathx.Primes(d.factor)
		lt       = len(t) - 1
		refreshC = time.NewTicker(time.Second * time.Duration(t[lt-2]))
		findNC   = time.NewTimer(time.Second * time.Duration(t[lt-1]))
	)
	defer refreshC.Stop()
	defer findNC.Stop()

	for {
		select {
		//退出事件
		case <-d.ctx.Done():
			goto loopRPCEnd

		//KClose事件
		case <-findNC.C:
			d.doKClose(d.ctx)

		//刷新事件
		case <-refreshC.C:
			d.doRefresh(d.ctx)

		}
	}

loopRPCEnd:
	d.logger.Infow(d.ctx, "dht srv main loop rpc end")

}

func (d *dht) doSend(ctx context.Context, pkt IDHTPacket) {
	var n = len(pkt.Body())

	if n > maxPacketSize {
		d.logger.Warnw(ctx, "dht srv message exceeds the maximum value", "cur", n, "max", maxPacketSize)
		return
	}

	nn, err := d.conn.WriteToUDP(pkt.Body(), pkt.Addr())
	if err != nil {
		d.logger.Warnw(ctx, "dht srv send message fail", "err", err)
		return
	}

	if nn < n {
		d.logger.Warnw(ctx, "dht srv message sent incomplete", "sent", nn, "should", n)
	}
}

func (d *dht) doReceive(ctx context.Context, pkt IDHTPacket) {
	var (
		msg = new(proto.Packet)
		err = enc.Unmarshal(pkt.Body(), msg)
	)
	if err != nil {
		d.logger.Warnw(ctx, "receive package", "type", errInvalidMessage, "err", err)
		return
	}

	var (
		sctx = context.WithValue(ctx, logx.TraceId, string(msg.GetTrace()))
	)
	switch msg.GetType() {
	case proto.PacketType_null:

	case proto.PacketType_ping:
		var ping = new(proto.Ping)
		err = enc.Unmarshal(msg.GetPkt(), ping)
		if err != nil {
			d.logger.Warnw(ctx, "receive ping message", "type", errInvalidMessage, "err", err)
			return
		}
		d.doPing(sctx, ping)

	case proto.PacketType_store:

	case proto.PacketType_findnode:
		var findnode = new(proto.FindNode)
		err = enc.Unmarshal(msg.GetPkt(), findnode)
		if err != nil {
			d.logger.Warnw(ctx, "receive findnode message", "type", errInvalidMessage, "err", err)
			return
		}
		d.doFindNode(sctx, findnode)

	case proto.PacketType_findvalue:

	}
}

func (d *dht) doFindNode(ctx context.Context, msg *proto.FindNode) {
	if msg.GetTo() != d.self.Encode() {
		return
	}
	var now = time.Now().UnixNano()

	switch msg.GetState() {
	case proto.State_ack:
		kclose := msg.GetK()
		for i := range kclose {
			d.store(ctx, d.self.Decode(kclose[i]))
		}

	case proto.State_syn:
		msg.State = proto.State_ack
		msg.Rt = now
		msg.To = msg.From
		msg.From = d.self.Encode()

		//query k-closest
		msg.K = d.closest(ctx, d.self.Decode(msg.Q), d.k)

		buf, err := enc.Marshal(msg)
		if err != nil {
			d.logger.Warnw(ctx, "dht srv marshal findnode message", "err", err)
			return
		}

		pkt := &proto.Packet{
			Type: proto.PacketType_findnode,
			Pkt:  buf,
		}

		d.send(ctx, d.self.Decode(msg.To), pkt)
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
			Pkt:  buf,
		}

		d.send(ctx, d.self.Decode(msg.To), pkt)
	}
}

func (d *dht) doRefresh(ctx context.Context) {
	d.rwm.RLock()
	defer d.rwm.RUnlock()

	var now = time.Now().UnixNano()
	var sctx = context.WithValue(ctx, logx.TraceId, mathx.UUID())

	for node, value := range d.nodes {
		var msg = new(proto.Ping)

		msg.From = d.self.Encode()
		msg.To = node
		msg.St = now
		msg.State = proto.State_syn

		buf, err := enc.Marshal(msg)
		if err != nil {
			d.logger.Warnw(sctx, "send ping message", "type", errInvalidMessage, "err", err, "msg", msg)
			return
		}

		pkt := &proto.Packet{
			Type: proto.PacketType_ping,
			Pkt:  buf,
		}

		d.send(ctx, d.self.Decode(node), pkt)
		value.total++

		if node == d.self.Encode() {
			d.local(ctx)
		}
		value.kpi()
	}
}

func (d *dht) doKClose(ctx context.Context) {

	var (
		now    = time.Now().UnixNano()
		sctx   = context.WithValue(ctx, logx.TraceId, mathx.UUID())
		kClose = d.closest(sctx, d.self, d.k)
	)

	for _, node := range kClose {
		var msg = new(proto.FindNode)

		msg.From = d.self.Encode()
		msg.To = node
		msg.St = now
		msg.State = proto.State_syn
		msg.Q = d.self.Encode()

		buf, err := enc.Marshal(msg)
		if err != nil {
			d.logger.Warnw(sctx, "send findnode message", "type", errInvalidMessage, "err", err, "msg", msg)
			return
		}

		pkt := &proto.Packet{
			Type: proto.PacketType_findnode,
			Pkt:  buf,
		}

		d.send(ctx, d.self.Decode(node), pkt)
	}
}

func (d *dht) doFix(ctx context.Context) {
	var sctx = context.WithValue(ctx, logx.TraceId, mathx.UUID())
	nodes := d.rtm.ListU64()

	var index = make(map[uint64]bool)

	for _, node := range nodes {
		index[node] = true
	}

	for node := range d.nodes {
		if _, ok := index[node]; ok {
			continue
		}
		err := d.rtm.AddU64(node, 0xffffffffffffffff, node)
		if err != nil {
			d.logger.Warnw(sctx, "dht srv add rtm fail", "err", err)
		}
	}

	for node := range index {
		if _, ok := d.nodes[node]; ok {
			continue
		}
		err := d.rtm.DelU64(node, 0xffffffffffffffff)
		if err != nil {
			d.logger.Warnw(sctx, "dht srv del rtm fail", "err", err)
		}
	}

	for i := range d.bucket {
		d.bucket[i] = 0
	}

	for node := range d.nodes {
		d.bucket[d.xor(d.self, d.self.Decode(node))]++
	}
}

func (d *dht) doEvolut(ctx context.Context) {
	var sctx = context.WithValue(ctx, logx.TraceId, mathx.UUID()) //优胜劣汰，适者生存
	var bad []uint64

	for node, value := range d.nodes {
		switch {
		case value.score > kpiJ:
			bad = append(bad, node)
		}
	}

	d.logger.Infow(ctx, "dht srv evolut bad node", "bad", bad)

	for i := range bad {
		node := d.self.Decode(bad[i])
		d.remove(sctx, node)
	}
}

func (d *dht) doGc(ctx context.Context) {
	//1. 回收Conn
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

	if trace, ok := ctx.Value(logx.TraceId).(string); ok {
		pkt.Trace = []byte(trace)
	} else {
		pkt.Trace = []byte(mathx.UUID())
	}

	buf, err := enc.Marshal(pkt)
	if err != nil {
		d.logger.Warnw(ctx, "send package", "type", errInvalidMessage, "err", err, "msg", pkt)
		return
	}

	select {
	case d.packetOutCh <- NewPacket(ctx, d.conn.Addr(to), buf):
	case <-d.ctx.Done():
		return
	}
}

func (d *dht) xor(src INode, dst INode) uint8 {
	var (
		dist  = d.bit - 1 //[0,d.bit)
		probe = uint64(0x8000000000000000)
		//距离的定义:
		// src: 101010101010110.........10101
		// dst: 101010101010111.........10101
		//  ^ : 000000000000001.........00000
		//		-----------------------------
		//	                  |--->dist<----|
		x = src.Encode() ^ dst.Encode()
	)

	for probe > 0 {
		if probe&x > 0 {
			break
		}
		probe >>= 1
		dist--
	}

	if dist > d.bit || dist < 0 { //溢出
		dist = 0
	}

	return dist % d.bit
}

func (d *dht) store(ctx context.Context, node INode) {
	if node == nil {
		return
	}
	var err = d.rtm.AddU64(node.Encode(), 0xffffffffffffffff, node.Encode())
	if err != nil {
		d.logger.Warnw(ctx, "dht srv store node fail", "err", err, "node", node.Encode())
		return
	}
	var kpi kpiL
	if node.Encode() == d.self.Encode() {
		kpi = kpiSSS
	} else {
		kpi = kpiZ
	}

	if x := d.xor(d.self, node); d.bucket[x] > d.k {
		d.logger.Infow(ctx, "dht srv bucket is full", "slot", x)
		return
	}

	d.rwm.Lock()
	d.nodes[node.Encode()] = newDHTNode(node, kpi)
	d.rwm.Unlock()
}

func (d *dht) remove(ctx context.Context, node INode) {
	if node == nil || node.Encode() == d.self.Encode() {
		return
	}

	var err = d.rtm.DelU64(node.Encode(), 0xffffffffffffffff)
	if err != nil {
		d.logger.Warnw(ctx, "dht srv remove node fail", "err", err, "node", node.Encode())
		return
	}

	d.rwm.Lock()
	delete(d.nodes, node.Encode())
	d.rwm.Unlock()
}

func (d *dht) seeds(ctx context.Context, nodes []uint64) {
	for i := range nodes {
		d.store(ctx, d.self.Decode(nodes[i]))
	}
}

func (d *dht) dump(ctx context.Context) interface{} {
	d.rwm.RLock()
	defer d.rwm.RUnlock()

	var r = make([]map[string]interface{}, 0, len(d.nodes))
	for node, value := range d.nodes {
		r = append(r, map[string]interface{}{
			"nodeId": fmt.Sprintf("0x%x", node),
			"score":  fmt.Sprintf("0x%x", value.score),
			"total":  value.total,
			"succ":   value.succ,
			"keep":   value.keep,
			"delay":  value.delay,
		})
	}
	return r
}

func (d *dht) closest(ctx context.Context, node INode, k uint8) []uint64 {
	var (
		dist  = make([]uint8, 0, d.bit)
		nodes = make([]uint64, 0, d.bit)
	)

	d.rwm.RLock()
	for n := range d.nodes {
		var x = d.xor(node, d.self.Decode(n))
		dist = append(dist, x)
		nodes = append(nodes, n)
	}
	d.rwm.RUnlock()

	var h, err = NewDHTHeap(dist, nodes)
	if err != nil {
		d.logger.Warnw(ctx, "dht srv k-close node fail", "err", err, "node", node.Encode())
		return nil
	}

	return h.Top(k)
}

func (d *dht) random(ctx context.Context) []uint64 {
	var r = make([]uint64, 0, d.alpha)

	d.rwm.RLock()
	defer d.rwm.RUnlock()

	for node := range d.nodes {
		if len(r) >= int(d.alpha) {
			break
		}
		r = append(r, node)
	}

	return r
}

func (d *dht) local(ctx context.Context) {
	d.rwm.RLock()
	defer d.rwm.RUnlock()

	if value, ok := d.nodes[d.self.Encode()]; ok {
		value.delay = 0
		value.succ++
	}
}
