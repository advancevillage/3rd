package dht

import (
	"container/heap"
	"errors"
)

var (
	errInvalidParam = errors.New("invalid param")
)

type dhtHeapNode struct {
	dist uint8
	node uint64
}

type IDHTHeap interface {
	Top(k uint8) []uint64
}

type DHTHeap []dhtHeapNode

func (h DHTHeap) Len() int {
	return len(h)
}

func (h DHTHeap) Less(i, j int) bool {
	return h[i].dist < h[j].dist
}

func (h DHTHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *DHTHeap) Push(a interface{}) {
	*h = append(*h, a.(dhtHeapNode))
}

func (h *DHTHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func (h *DHTHeap) Top(k uint8) []uint64 {
	var top = make([]uint64, 0, k)

	for k > 0 && h.Len() > 0 {
		v := heap.Pop(h).(dhtHeapNode)
		top = append(top, v.node)
		k--
	}

	return top
}

func NewDHTHeap(dist []uint8, nodes []uint64) (IDHTHeap, error) {
	if len(dist) != len(nodes) {
		return nil, errInvalidParam
	}
	var h = new(DHTHeap)
	var l = len(dist)

	for i := 0; i < l; i++ {
		heap.Push(h, dhtHeapNode{dist: dist[i], node: nodes[i]})
	}
	return h, nil
}
