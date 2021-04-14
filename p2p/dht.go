package p2p

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/advancevillage/3rd/database"
	"github.com/advancevillage/3rd/ecies"
)

type IDHT interface {
	Start()
}

const (
	nBytes   = 32
	nBucket  = nBytes * 8
	nLimit   = 20
	conLimit = 50

	tableName = "dht"
	enodeKey  = "enode:"

	refreshT = 10
	liveT    = 30
	storeT   = 60
)

//@overview: kad table
//@author: richard.sun
//@param:
//1. 核心是确定xor索引方式
type bucket struct {
	nodes []*node
}

func (b *bucket) set(enode ecies.IENode) {
	//1. 容量预检
	if len(b.nodes) > nLimit {
		return
	}
	b.nodes = append(b.nodes, &node{enode: enode, ts: time.Now().Unix()})
}

func (b *bucket) update(enode ecies.IENode) {
	var flag = false
	for _, v := range b.nodes {
		if bytes.Equal(enode.GetId(), v.enode.GetId()) {
			v.ts = time.Now().Unix()
			flag = true
		} else {
			continue
		}
	}
	if !flag {
		b.nodes = append(b.nodes, &node{enode: enode, ts: time.Now().Unix()})
	}
}

type node struct {
	enode ecies.IENode
	ts    int64 //unix timestamp
}
type dht struct {
	db       database.ILevelDB //存储
	boot     []ecies.IENode    //启动节点
	local    ecies.IENode      //本地节点
	table    [nBucket]*bucket  //分布式Hash表
	conLimit int               //并发控制

	udpq IP2PQueue
	dhtq IP2PQueue
	ctx  context.Context
}

func NewDHT(ctx context.Context, boot []ecies.IENode, local ecies.IENode, udpq IP2PQueue, dhtq IP2PQueue) (IDHT, error) {
	//1. 创建DHT对象
	var d = &dht{}
	var err error
	d.boot = boot
	d.dhtq = dhtq
	d.udpq = udpq
	d.local = local
	d.ctx = ctx
	d.conLimit = conLimit
	//2. 初始化 存储
	d.db, err = database.NewPersistentStore(tableName)
	if err != nil {
		return nil, fmt.Errorf("create dht table database err. %s", err.Error())
	}
	//3. 初始化 DHT表
	err = d.loadDHTTable(append(boot, local)...)
	if err != nil {
		return nil, fmt.Errorf("load dht table err. %s", err.Error())
	}
	return d, nil
}

//@overview: 将节点信息从数据加载出来
//@param:
//1. boot 启动节点载入DHT表
func (d *dht) loadDHTTable(boot ...ecies.IENode) error {
	//1. 初始化表
	var err error
	var n int
	for i := 0; i < nBucket; i++ {
		var bkt = &bucket{}
		bkt.nodes = make([]*node, 0, nLimit)
		d.table[i] = bkt
	}
	//2. 载入启动节点
	//3. 载入存储节点
	var nodes = d.query()
	boot = append(boot, nodes...)
	for _, v := range boot {
		n, err = d.dist(d.local, v)
		if err != nil {
			continue
		}
		n %= nBucket
		d.table[n].set(v)
	}
	return nil
}

func (d *dht) dist(local ecies.IENode, remote ecies.IENode) (int, error) {
	var src = local.GetId()
	var dst = remote.GetId()
	var i = 0
	var j = 0
	if len(src) != len(dst) && len(src) != nBytes {
		return i, fmt.Errorf("local id len %d != %d != %d remote id len", len(src), nBytes, len(dst))
	}
	for i = 0; i < nBytes; i++ {
		src[i] ^= dst[i]
	}
	for i = 0; i < nBytes && src[i] <= 0; i++ {
	}
	i %= nBytes
	var di = src[i]
	for j = 0; j < 8 && (di>>j) > 0; j++ {
	}
	j %= 8
	return i*8 + j, nil
}

func (d *dht) query() []ecies.IENode {
	var nodes []ecies.IENode
	var m, err = d.db.Range([]byte(enodeKey))
	if err != nil {
		return nodes
	}
	for k, v := range m {
		var enode ecies.IENode
		enode, err = ecies.NewENodeByUrl(v)
		if err != nil {
			continue
		}
		if k[len(enodeKey):] != string(enode.GetId()) {
			continue
		}
		nodes = append(nodes, enode)
	}
	return nodes
}

func (d *dht) store(nodes []ecies.IENode) {
	for _, v := range nodes {
		var value, err = v.GetENodeUrl()
		if err != nil {
			continue
		}
		var key = fmt.Sprintf("%s%s", enodeKey, string(v.GetId()))
		err = d.db.Put([]byte(key), []byte(value))
		if err != nil {
			continue
		}
	}
}

func (d *dht) refresh() {
	var conLimit = make(chan struct{}, d.conLimit)
	for _, bkt := range d.table {
		for _, n := range bkt.nodes {
			conLimit <- struct{}{}
			go func(nn *node) {
				defer func() {
					<-conLimit
				}()
				var m = newmsg(TypePing, nn.enode.GetId(), nn.enode)
				var t = time.NewTicker(time.Second * 3)
				d.udpq.Push(m)
				select {
				case <-m.out:
					nn.ts = time.Now().Unix()
				case <-t.C:
				}
			}(n)
			conLimit <- struct{}{}
			go func(nn *node) {
				defer func() {
					<-conLimit
				}()
				var m = newmsg(TypeFindNode, d.local.GetId(), nn.enode)
				var t = time.NewTicker(time.Second * 3)
				d.udpq.Push(m)
				select {
				case b := <-m.out:
					var ks = strings.Split(string(b), "|")
					for _, v := range ks {
						var enode, err = ecies.NewENodeByUrl(v)
						if err != nil {
							continue
						}
						ix, err := d.dist(d.local, enode)
						if err != nil {
							continue
						}
						d.table[ix].update(enode)
					}
				case <-t.C:
				}
			}(n)
		}
	}
}

func (d *dht) persistent() {
	//TODO: 删除历史Key
	for i := 0; i < len(d.table); i++ {
		for _, n := range d.table[i].nodes {
			var key = fmt.Sprintf("%s%s", enodeKey, string(n.enode.GetId()))
			var err = d.db.Put([]byte(key), n.enode.GetId())
			if err != nil {
				continue
			}
		}
	}
}

func (d *dht) live() {
	for i := 0; i < len(d.table); i++ {
		var bkt = d.table[i]
		var now = time.Now().Unix()
		for j := 0; j < len(bkt.nodes); j++ {
			if now-bkt.nodes[j].ts > 30 {
				bkt.nodes = append(bkt.nodes[:j], bkt.nodes[j+1:]...)
				j--
			}
		}
	}
}

func (d *dht) kClosest(enode ecies.IENode) []byte {
	var kClosest = make([]ecies.IENode, 0, nLimit)
	kClosest = append(kClosest, d.local)
	for _, bkt := range d.table {
		for _, n := range bkt.nodes {
			ix := sort.Search(len(kClosest), func(i int) bool {
				var da, e1 = d.dist(n.enode, enode)
				var db, e2 = d.dist(kClosest[i], enode)
				if e1 != nil || e2 != nil {
					return false
				}
				return db < da
			})
			if len(kClosest) < nLimit {
				kClosest = append(kClosest, n.enode)
			}
			if ix >= nLimit {
				continue
			} else {
				copy(kClosest[ix+1:], kClosest[ix:])
				kClosest[ix] = n.enode
			}
		}
	}
	var m = make([]string, 0, nLimit)
	for _, v := range kClosest {
		m = append(m, string(v.GetId()))
	}
	var s = strings.Join(m, "|")
	return []byte(s)
}

func (d *dht) closest(enode ecies.IENode) []byte {
	var closest = d.local
	for _, bkt := range d.table {
		for _, n := range bkt.nodes {
			var da, e1 = d.dist(closest, enode)
			var db, e2 = d.dist(n.enode, enode)
			if e1 != nil || e2 != nil {
				continue
			}
			if db < da {
				closest = n.enode
			}
		}
	}
	return closest.GetId()
}

func (d *dht) handle(m *msg) {
	//1. 参数检查
	switch {
	case nil == m:
	case nil == m.out:
	case nil == m.in:
	case 0 >= len(m.in):
	case TypePing == m.tYpe:
		var enode, err = ecies.NewENodeByUrl(string(m.in))
		if err != nil {
			return
		}
		m.out <- []byte{}
		ix, err := d.dist(d.local, enode)
		if err != nil {
			return
		}
		d.table[ix].update(enode)
	case TypeFindNode == m.tYpe:
		var enode, err = ecies.NewENodeByUrl(string(m.in))
		if err != nil {
			return
		}
		m.out <- d.kClosest(enode)
	case TypeFindValue == m.tYpe:
		var enode, err = ecies.NewENodeByUrl(string(m.in))
		if err != nil {
			return
		}
		m.out <- d.closest(enode)
	default:
	}
}

func (d *dht) loop() {
	var (
		rT = time.NewTicker(time.Second * refreshT)
		lT = time.NewTicker(time.Second * liveT)
		sT = time.NewTicker(time.Second * storeT)
	)
	for {
		select {
		case <-rT.C:
			d.refresh()
		case <-sT.C:
			d.persistent()
		case <-lT.C:
			d.live()
		//事件源来自udp事件队列
		case m := <-d.dhtq.Pull():
			go d.handle(m)
		}
	}
}

func (d *dht) Start() {
	go d.loop()
}
