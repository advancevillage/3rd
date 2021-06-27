package p2p

import (
	"context"
	"fmt"
	"time"

	"github.com/advancevillage/3rd/ecies"
	"github.com/advancevillage/3rd/monitor"
)

type IP2PQueue interface {
	monitor.IMonitor
	Push(*msg)
	Pull() <-chan *msg
}

type queue struct {
	ch      chan *msg
	timeout time.Duration
}

type msg struct {
	out   chan []byte
	in    []byte
	tYpe  byte
	enode ecies.IENode
}

func newmsg(tYpe byte, in []byte, enode ecies.IENode) *msg {
	return &msg{
		tYpe:  tYpe,
		in:    in,
		out:   make(chan []byte, 1),
		enode: enode,
	}
}

func (q *queue) Monitor() interface{} {
	var m = make(map[string]interface{})
	m["timeout"] = q.timeout
	m["chLen"] = len(q.ch)
	m["chCap"] = cap(q.ch)
	return m
}

func newQueue() IP2PQueue {
	return &queue{
		ch:      make(chan *msg, 1024),
		timeout: time.Second * 2,
	}
}

func (q *queue) Push(m *msg) {
	var t = time.NewTicker(q.timeout)
	select {
	case q.ch <- m:
	case <-t.C:
	}
}

func (q *queue) Pull() <-chan *msg {
	return q.ch
}

type IP2P interface {
	monitor.IMonitor
	Start()
}

type p2P struct {
	disc IKad
	dht  IDHT
	ctx  context.Context
	quit context.CancelFunc
}

func NewP2P(local ecies.IENode, boot []ecies.IENode) (IP2P, error) {
	var p = &p2P{}
	var err error
	var udpq = newQueue()
	var dhtq = newQueue()
	p.ctx, p.quit = context.WithCancel(context.TODO())
	p.dht, err = NewDHT(p.ctx, boot, local, udpq, dhtq)
	if err != nil {
		return nil, fmt.Errorf("new dht client fail. err %s", err.Error())
	}
	p.disc, err = NewKad(p.ctx, local, udpq, dhtq)
	if err != nil {
		return nil, fmt.Errorf("new kad client fail. err %s", err.Error())
	}
	return p, nil
}

func (p *p2P) Start() {
	p.dht.Start()
	p.disc.Start()
}

func (p *p2P) Monitor() interface{} {
	var m = make(map[string]interface{})
	m["disc"] = p.disc.Monitor()
	m["dht"] = p.dht.Monitor()
	return m
}
