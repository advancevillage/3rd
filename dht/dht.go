package dht

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/advancevillage/3rd/logx"
	"github.com/advancevillage/3rd/mathx"
	"github.com/advancevillage/3rd/proto"
	enc "github.com/golang/protobuf/proto"
)

var (
	errInvalidHost     = errors.New("invalid host")
	errInvalidPort     = errors.New("invalid port")
	errInvalidNode     = errors.New("invalid node")
	errInvalidMessage  = errors.New("invalid message")
	errInvalidProtocol = errors.New("invalid protocol")

	maxPacketSize = 1024
)

type IDHT interface {
	Start()
}

type DHTCfg struct {
	Refresh int      `json:"refresh"`
	Evolut  int      `json:"evolut"`
	Zone    uint16   `json:"zone"`
	Network string   `json:"network"`
	Addr    string   `json:"addr"`
	Seeds   []uint64 `json:"seeds"`
	Alpha   int      `json:"alpha"`
	K       int      `json:"k"`
}

func (d *dht) Start() {
	go d.readLoop()
	go d.loop()

	select {
	case <-d.ctx.Done():
		d.logger.Infow(d.ctx, "dht srv exit")
	}
}

func NewDHT(ctx context.Context, logger logx.ILogger, cfg *DHTCfg) (IDHT, error) {
	node, err := NewNodeWithAddr(cfg.Zone, cfg.Addr)
	if err != nil {
		return nil, err
	}
	conn, err := NewConn(cfg.Network, node)
	if err != nil {
		return nil, err
	}

	var d = &dht{}

	d.ctx, d.canecl = context.WithCancel(ctx)
	d.logger = logger
	d.cfg = cfg
	d.nodeCli = newBktMgr(ctx, logger, node, 4, 64, cfg.Seeds)
	d.connCli = conn
	d.packetInCh = make(chan IDHTPacket, cfg.Alpha)
	d.packetOutCh = make(chan IDHTPacket, cfg.Alpha)

	return d, nil
}

type dht struct {
	wg          sync.WaitGroup
	ctx         context.Context
	canecl      context.CancelFunc
	logger      logx.ILogger
	cfg         *DHTCfg  //配置管理
	nodeCli     IDHTNode //节点管理
	connCli     IDHTConn //传输管理
	packetInCh  chan IDHTPacket
	packetOutCh chan IDHTPacket
}

func (d *dht) loop() {
	go d.loopNetIO()
	go d.loopRefresh()
	d.wg.Wait()
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

func (d *dht) doSend(ctx context.Context, pkt IDHTPacket) {
	var n = len(pkt.Body())

	if n > maxPacketSize {
		d.logger.Warnw(ctx, "dht srv message exceeds the maximum value", "cur", n, "max", maxPacketSize)
		return
	}

	nn, err := d.connCli.WriteToUDP(pkt.Body(), pkt.Addr())
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

	case proto.PacketType_findvalue:

	}
}

func (d *dht) loopRefresh() {
	d.wg.Add(1)
	defer d.wg.Done()

	var refreshC = time.NewTicker(time.Second * time.Duration(d.cfg.Refresh))
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

func (d *dht) doRefresh(ctx context.Context) {
	var (
		wg    sync.WaitGroup
		connC = make(chan struct{}, d.cfg.Alpha)
		now   = time.Now().UnixNano()
		sctx  = context.WithValue(ctx, logx.TraceId, mathx.UUID())
		nodes = d.nodeCli.List(sctx)
	)

	for _, v := range nodes {
		wg.Add(1)
		connC <- struct{}{}

		go func(node INode) {
			defer func() {
				wg.Done()
				<-connC
			}()

			var msg = new(proto.Ping)
			msg.From = d.nodeCli.Self(sctx)
			msg.To = Encode(node)
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
			d.send(sctx, node, pkt)
			d.nodeCli.Syn(sctx, node)
		}(v)
	}
	wg.Wait()
}

func (d *dht) doPing(ctx context.Context, msg *proto.Ping) {
	//1. 检查目的节点是不是给自己的
	if msg.GetTo() != d.nodeCli.Self(ctx) {
		return
	}
	var now = time.Now().UnixNano()
	//2. 检查消息状态
	switch msg.GetState() {
	case proto.State_ack:
		d.nodeCli.Ack(ctx, Decode(msg.GetFrom()), now-msg.GetSt())

	case proto.State_syn:
		msg.State = proto.State_ack
		msg.Rt = now
		msg.To = msg.From
		msg.From = d.nodeCli.Self(ctx)

		buf, err := enc.Marshal(msg)
		if err != nil {
			d.logger.Warnw(ctx, "dht srv marshal ping message", "err", err)
			return
		}

		pkt := &proto.Packet{
			Type: proto.PacketType_ping,
			Pkt:  buf,
		}

		d.send(ctx, Decode(msg.GetTo()), pkt)
	}
}

func (d *dht) readLoop() {
	for {
		buf, from, err := d.connCli.ReadFromUDP()

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
	case d.packetOutCh <- NewPacket(ctx, d.connCli.Addr(to), buf):
	case <-d.ctx.Done():
		return
	}
}

type IDHTNode interface {
	Self(ctx context.Context) uint64
	List(ctx context.Context) []INode
	Syn(ctx context.Context, n INode)
	Ack(ctx context.Context, n INode, delay int64)
	Set(ctx context.Context, n INode)
	Del(ctx context.Context, n INode)
	Show(ctx context.Context) interface{}
}

type bktmgr struct {
	bit    int
	self   INode
	bkts   []IDHTBkt
	logger logx.ILogger
}

func newBktMgr(ctx context.Context, logger logx.ILogger, self INode, k int, bit int, seeds []uint64) IDHTNode {
	var mgr = new(bktmgr)
	mgr.bkts = make([]IDHTBkt, bit)
	mgr.self = self
	mgr.bit = bit
	mgr.logger = logger

	for i := 0; i < bit; i++ {
		mgr.bkts[i] = newBkt(logger, k)
	}

	seeds = append(seeds, Encode(self))

	for i := range seeds {
		mgr.Set(ctx, Decode(seeds[i]))
	}

	return mgr
}

func (mgr *bktmgr) Self(ctx context.Context) uint64 {
	return Encode(mgr.self)
}

func (mgr *bktmgr) List(ctx context.Context) []INode {
	var n []INode

	for i := 0; i < mgr.bit; i++ {
		var nn = mgr.bkts[i].List(ctx)
		n = append(n, nn...)
	}
	mgr.Show(ctx)
	return n
}

func (mgr *bktmgr) Syn(ctx context.Context, n INode) {
	mgr.logger.Infow(ctx, fmt.Sprintf("bucket[%d] syn node", XOR(mgr.self, n)), "node", Encode(n))
	mgr.bkts[XOR(mgr.self, n)].Syn(ctx, n)
}

func (mgr *bktmgr) Ack(ctx context.Context, n INode, delay int64) {
	mgr.logger.Infow(ctx, fmt.Sprintf("bucket[%d] ack node", XOR(mgr.self, n)), "node", Encode(n))
	mgr.bkts[XOR(mgr.self, n)].Ack(ctx, n, delay)
}

func (mgr *bktmgr) Set(ctx context.Context, n INode) {
	mgr.logger.Infow(ctx, fmt.Sprintf("bucket[%d] set node", XOR(mgr.self, n)), "node", Encode(n))
	mgr.bkts[XOR(mgr.self, n)].Set(ctx, n)
}

func (mgr *bktmgr) Del(ctx context.Context, n INode) {
	mgr.logger.Infow(ctx, fmt.Sprintf("bucket[%d] del node", XOR(mgr.self, n)), "node", Encode(n))
	mgr.bkts[XOR(mgr.self, n)].Del(ctx, n)
}

func (mgr *bktmgr) Show(ctx context.Context) interface{} {
	var r []interface{}

	for i := 0; i < mgr.bit; i++ {
		var rr = mgr.bkts[i].Show(ctx)
		r = append(r, rr)
	}
	mgr.logger.Infow(ctx, "show bucket[:]", "info", r)
	return r

}

type IDHTBkt interface {
	Len() int
	Evolut(ctx context.Context)
	List(ctx context.Context) []INode
	Syn(ctx context.Context, n INode)
	Ack(ctx context.Context, n INode, delay int64)
	Set(ctx context.Context, n INode)
	Del(ctx context.Context, n INode)
	Show(ctx context.Context) interface{}
}

type bucket struct {
	k      int
	rwl    sync.RWMutex
	nodes  map[uint64]*node
	logger logx.ILogger
}

func newBkt(logger logx.ILogger, k int) IDHTBkt {
	return &bucket{
		nodes:  make(map[uint64]*node),
		k:      k,
		logger: logger,
	}
}

func (b *bucket) Len() int {
	return len(b.nodes)
}

func (b *bucket) List(ctx context.Context) []INode {
	b.rwl.RLock()
	defer b.rwl.RUnlock()

	var l = make([]INode, 0, b.Len())

	for k := range b.nodes {
		l = append(l, Decode(k))
	}
	return l
}

func (b *bucket) Syn(ctx context.Context, node INode) {
	b.rwl.RLock()
	var n, ok = b.nodes[Encode(node)]
	b.rwl.RUnlock()

	if !ok {
		return
	}

	n.fail++
	n.kpi()
}

func (b *bucket) Ack(ctx context.Context, node INode, delay int64) {
	b.rwl.RLock()
	var n, ok = b.nodes[Encode(node)]
	b.rwl.RUnlock()

	if !ok {
		return
	}

	n.delay = delay
	n.fail--
}

func (b *bucket) Set(ctx context.Context, node INode) {
	if b.Len() > b.k {
		return
	}

	b.rwl.Lock()
	b.nodes[Encode(node)] = newNode(node)
	b.rwl.Unlock()
}

func (b *bucket) Del(ctx context.Context, node INode) {
	b.rwl.Lock()
	delete(b.nodes, Encode(node))
	b.rwl.Unlock()

	b.logger.Infow(ctx, "bucket del node", "node", node)
}

func (b *bucket) Show(ctx context.Context) interface{} {
	b.rwl.RLock()
	defer b.rwl.RUnlock()

	var r = make([]map[string]interface{}, 0, b.Len())
	for node, value := range b.nodes {
		r = append(r, map[string]interface{}{
			"id":    fmt.Sprintf("0x%x", value.id.Id()),
			"enode": fmt.Sprintf("0x%x", node),
			"score": fmt.Sprintf("0x%x", value.score),
			"delay": value.delay,
			"fail":  value.fail,
			"keep":  value.keep,
		})
	}
	return r
}

func (b *bucket) Evolut(ctx context.Context) {
	var bad []uint64

	b.rwl.RLock()
	for k, v := range b.nodes {
		if v.score > kpiO {
			bad = append(bad, k)
		} else {
			continue
		}
	}
	b.rwl.RUnlock()

	for i := range bad {
		b.Del(ctx, Decode(bad[i]))
	}
}

func newNode(id INode) *node {
	return &node{
		id:    id,
		at:    time.Now(),
		delay: 0,
		fail:  0,
		keep:  0,
		score: kpiZ,
	}
}

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

type node struct {
	id    INode     //节点
	at    time.Time //节点加入时间
	delay int64     //纳秒
	fail  int64
	keep  int64
	score kpiL
}

func (n *node) kpi() {
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
