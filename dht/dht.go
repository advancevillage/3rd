package dht

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/advancevillage/3rd/logx"
	"github.com/advancevillage/3rd/radix"
)

var (
	errInvalidHost = errors.New("invalid host")
	errInvalidPort = errors.New("invalid port")
)

type IDHT interface {
}

func NewDHT(ctx context.Context, logger logx.ILogger, zone uint16, network string, addr string) (IDHT, error) {
	var d = &dht{}
	d.ctx = ctx
	d.logger = logger

	d.k = 4
	d.bit = 0x40
	d.alpha = 3
	d.bucket = make([]uint8, d.bit)

	d.rtm = radix.NewRadixTree()
	d.nodes = make(map[uint64]*dhtnode)

	d.refresh = 10
	d.quit = make(chan bool, 1)

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

	d.self = NewNode(network, zone, uint16(port), ipv4)

	return d, nil
}

type dht struct {
	self   INode
	ctx    context.Context
	logger logx.ILogger

	//常数设置
	k      uint8
	alpha  uint8
	bit    uint8
	bucket []uint8

	//路由表
	rtm   radix.IRadixTree
	nodes map[uint64]*dhtnode //存储节点列表
	rwm   sync.RWMutex

	//事件指令
	refresh int
	quit    chan bool
}

type dhtnode struct {
	node  INode     //节点
	at    time.Time //节点加入时间
	score uint32    //节点评分 score = latencyScore + succScore
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
		}
	}

end:
	d.doGc()
}

func (d *dht) doRefresh() {

}

func (d *dht) doFix() {

}

func (d *dht) doGc() {

}

func (d *dht) listen() {
	switch d.self.Protocol() {
	case "udp":
		d.listenUDP()
	default:
		d.logger.Errorw(d.ctx, "dht srv listen fail", "protocol", d.self.Protocol())
		d.quit <- true
	}
}

func (s *dht) listenUDP() {
	var a = byte(s.self.Ipv4() >> 24)
	var b = byte(s.self.Ipv4() >> 16)
	var c = byte(s.self.Ipv4() >> 8)
	var d = byte(s.self.Ipv4())

	var ip = net.IPv4(a, b, c, d)
	var port = int(s.self.Port())

	var conn, err = net.ListenUDP("udp", &net.UDPAddr{IP: ip, Port: port})
	if err != nil {
		s.logger.Errorw(s.ctx, "dht srv listen fail", "err", err)
		s.quit <- true
		return
	}
	defer conn.Close()
	s.logger.Infow(s.ctx, "dht srv listen", "addr", fmt.Sprintf("dht://%s(%d.%d.%d.%d:%d)", s.self.Protocol(), a, b, c, d, port))

	//TODO: 待拆分
	var rx = make([]byte, 1024)
	for {
		var rn, addr, err = conn.ReadFromUDP(rx)
		if err != nil {
			s.logger.Warnw(s.ctx, "read from udp", "addr", addr, "err", err)
			continue
		}
		_, err = conn.WriteToUDP(rx[:rn], addr)
		if err != nil {
			s.logger.Warnw(s.ctx, "write to udp", "addr", addr, "err", err)
		}
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
