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
	kpiSSS = kpiL(0x01)
	kpiSS  = kpiL(0x02)
	kpiS   = kpiL(0x03)
	kpiA   = kpiL(0x04)
	kpiB   = kpiL(0x05)
	kpiC   = kpiL(0x06)
	kpiD   = kpiL(0x07)
	kpiE   = kpiL(0x08)
	kpiF   = kpiL(0x09)
	kpiG   = kpiL(0x0a)
	kpiH   = kpiL(0x0b)
	kpiI   = kpiL(0x0c)
	kpiJ   = kpiL(0x0e)
	kpiK   = kpiL(0x0f)
	kpiM   = kpiL(0x10)
	kpiN   = kpiL(0x11)
	kpiO   = kpiL(0x12)
	kpiP   = kpiL(0x13)
	kpiQ   = kpiL(0x14)
	kpiR   = kpiL(0x15)
	kpiT   = kpiL(0x16)
	kpiU   = kpiL(0x17)
	kpiV   = kpiL(0x18)
	kpiW   = kpiL(0x19)
	kpiX   = kpiL(0x1a)
	kpiY   = kpiL(0x1b)
	kpiZ   = kpiL(0xff)
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
	fail  int64
	keep  int64
	score kpiL
}

func newDHTNode(node INode, initScore kpiL) *dhtNode {
	return &dhtNode{
		node:  node,
		at:    time.Now(),
		delay: 0,
		fail:  0,
		keep:  0,
		score: initScore,
	}
}

//5 10 50 100 200 300 400 500 600 700 800 900 1000 1500 2000 2500 3000 3500 4000 5000 6000 7000  8000 9000 10000
//4 5  6   7   8   9   10  11 12   13  14  15  16   17   18   19   20   21   22   23   24   25    26   27   28
func (n *dhtNode) kpi() {
	var delay = n.delay / 1e6
	var fail = n.fail
	var last = n.score

	switch {
	case delay <= 5 && fail <= 4:
		n.score = kpiSSS
	case delay <= 10 && fail <= 5:
		n.score = kpiSS
	case delay <= 50 && fail <= 6:
		n.score = kpiS
	case delay <= 100 && fail <= 7:
		n.score = kpiA
	case delay <= 200 && fail <= 8:
		n.score = kpiB
	case delay <= 300 && fail <= 9:
		n.score = kpiC
	case delay <= 400 && fail <= 10:
		n.score = kpiD
	case delay <= 500 && fail <= 11:
		n.score = kpiE
	case delay <= 600 && fail <= 12:
		n.score = kpiF
	case delay <= 700 && fail <= 13:
		n.score = kpiG
	case delay <= 800 && fail <= 14:
		n.score = kpiH
	case delay <= 900 && fail <= 15:
		n.score = kpiI
	case delay <= 1000 && fail <= 16:
		n.score = kpiJ
	case delay <= 1500 && fail <= 17:
		n.score = kpiK
	case delay <= 2000 && fail <= 18:
		n.score = kpiM
	case delay <= 2500 && fail <= 19:
		n.score = kpiN
	case delay <= 3000 && fail <= 20:
		n.score = kpiO
	case delay <= 3500 && fail <= 21:
		n.score = kpiP
	case delay <= 4000 && fail <= 22:
		n.score = kpiQ
	case delay <= 5000 && fail <= 23:
		n.score = kpiR
	case delay <= 5000 && fail <= 24:
		n.score = kpiT
	case delay <= 6000 && fail <= 25:
		n.score = kpiU
	case delay <= 7000 && fail <= 26:
		n.score = kpiV
	case delay <= 8000 && fail <= 27:
		n.score = kpiW
	case delay <= 9000 && fail <= 28:
		n.score = kpiX
	case delay <= 10000 && fail <= 29:
		n.score = kpiY
	default:
		n.score = kpiZ
	}

	if last != n.score || n.score == kpiZ {
		n.keep = 0
		return
	}

	if p := n.keep / 0x10; p > 0 {
		n.fail -= p
		n.keep %= 0x10
	} else {
		n.keep++
	}
}

func (d *dht) loop() {
	go d.loopNetIO()
	go d.loopFix()
	go d.loopEvolut()
	go d.loopRefresh()
	go d.loopKClose()
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
			goto loopNetIOEnd
		}
	}

loopNetIOEnd:
	d.logger.Infow(d.ctx, "dht srv main loop net io end")

}

func (d *dht) loopFix() {
	d.wg.Add(1)
	defer d.wg.Done()

	var (
		t    = mathx.Primes(d.factor)
		lt   = len(t) - 1
		fixC = time.NewTicker(time.Second * time.Duration(t[lt-3]))
	)
	defer fixC.Stop()

	for {
		select {
		//退出事件
		case <-d.ctx.Done():
			goto loopFixEnd

		//修复事件
		case <-fixC.C:
			d.doFix(d.ctx)

		}
	}

loopFixEnd:
	d.logger.Infow(d.ctx, "dht srv main loop fix end")
}

func (d *dht) loopEvolut() {
	d.wg.Add(1)
	defer d.wg.Done()

	var (
		t       = mathx.Primes(d.factor)
		lt      = len(t) - 1
		evolutC = time.NewTicker(time.Second * time.Duration(t[lt]))
	)
	defer evolutC.Stop()

	for {
		select {
		//退出事件
		case <-d.ctx.Done():
			goto loopEvolutEnd

		//进化事件
		case <-evolutC.C:
			d.doEvolut(d.ctx)
		}
	}

loopEvolutEnd:
	d.logger.Infow(d.ctx, "dht srv main loop evolut end")
}

func (d *dht) loopRefresh() {
	d.wg.Add(1)
	defer d.wg.Done()

	var (
		t        = mathx.Primes(d.factor)
		lt       = len(t) - 1
		refreshC = time.NewTicker(time.Second * time.Duration(t[lt-2]))
	)
	defer refreshC.Stop()

	for {
		select {
		//退出事件
		case <-d.ctx.Done():
			goto loopRefreshEnd

		//刷新事件
		case <-refreshC.C:
			d.doRefresh(d.ctx)

		}
	}

loopRefreshEnd:
	d.logger.Infow(d.ctx, "dht srv main loop refresh end")

}

func (d *dht) loopKClose() {
	d.wg.Add(1)
	defer d.wg.Done()

	var (
		t      = mathx.Primes(d.factor)
		lt     = len(t) - 1
		findNC = time.NewTicker(time.Second * time.Duration(t[lt-1]))
	)
	defer findNC.Stop()

	for {
		select {
		//退出事件
		case <-d.ctx.Done():
			goto loopKCloseEnd

		//KClose事件
		case <-findNC.C:
			d.doKClose(d.ctx)

		}
	}

loopKCloseEnd:
	d.logger.Infow(d.ctx, "dht srv main loop kclose end")

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
		value.fail--
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
		value.fail++
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
	for {
		buf, from, err := d.conn.ReadFromUDP()

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
		case d.packetInCh <- NewPacket(d.ctx, from, buf):
		case <-d.ctx.Done():
			goto readLoopEnd
		}
	}

readLoopEnd:
	d.logger.Infow(d.ctx, "dnt srv readLoop end")
}

func (d *dht) send(ctx context.Context, to INode, pkt *proto.Packet) {
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

	d.rwm.RLock()
	var _, ok = d.nodes[node.Encode()]
	d.rwm.RUnlock()

	if ok {
		return
	}

	var err = d.rtm.AddU64(node.Encode(), 0xffffffffffffffff, node.Encode())
	if err != nil {
		d.logger.Warnw(ctx, "dht srv store node fail", "err", err, "node", node.Encode())
		return
	}

	if x := d.xor(d.self, node); d.bucket[x] > d.k {
		d.logger.Infow(ctx, "dht srv bucket is full", "slot", x)
		return
	}

	d.rwm.Lock()
	d.nodes[node.Encode()] = newDHTNode(node, kpiZ)
	d.rwm.Unlock()
}

func (d *dht) remove(ctx context.Context, node INode) {
	if node == nil {
		return
	}

	d.rwm.RLock()
	var _, ok = d.nodes[node.Encode()]
	d.rwm.RUnlock()

	if !ok {
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
			"delay":  value.delay,
			"fail":   value.fail,
			"keep":   value.keep,
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
