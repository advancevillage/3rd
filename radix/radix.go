package radix

import (
	"errors"
	"sync"
)

var (
	u32Empty = uint32(0xffffffff)
	u32Mask  = uint32(0xffffffff)
	u64Mask  = uint64(0xffffffffffffffff)
	u64Empty = uint64(0xffffffffffffffff)
	notFound = errors.New("not found")
)

type radixU32Node struct {
	l *radixU32Node //左节点
	r *radixU32Node //右节点
	p *radixU32Node //父节点
	v uint32
}

type radixU64Node struct {
	l *radixU64Node //左节点
	r *radixU64Node //右节点
	p *radixU64Node //父节点
	v uint64
}

func newRadixU64Node() *radixU64Node {
	var n = new(radixU64Node)

	n.r = nil
	n.l = nil
	n.p = nil
	n.v = u64Empty

	return n
}

func newRadixU32Node() *radixU32Node {
	var n = new(radixU32Node)

	n.r = nil
	n.l = nil
	n.p = nil
	n.v = u32Empty

	return n
}

func (rn *radixU32Node) isLeaf() bool {
	return rn.l == nil && rn.r == nil
}

func (rn *radixU32Node) isEmpty() bool {
	return rn.v == u32Empty
}

func (rn *radixU32Node) reset() {
	rn.r = nil
	rn.l = nil
	rn.p = nil
	rn.v = u32Empty
}

func (rn *radixU32Node) setL(l *radixU32Node) {
	rn.l = l
}

func (rn *radixU32Node) setR(r *radixU32Node) {
	rn.r = r
}

func (rn *radixU32Node) setP(p *radixU32Node) {
	rn.p = p
}

func (rn *radixU32Node) setV(v uint32) {
	rn.v = v
}

func (rn *radixU32Node) getL() *radixU32Node {
	return rn.l
}

func (rn *radixU32Node) getR() *radixU32Node {
	return rn.r
}

func (rn *radixU32Node) getP() *radixU32Node {
	return rn.p
}

func (rn *radixU32Node) getV() uint32 {
	return rn.v
}

func (rn *radixU64Node) isLeaf() bool {
	return rn.l == nil && rn.r == nil
}

func (rn *radixU64Node) isEmpty() bool {
	return rn.v == u64Empty
}

func (rn *radixU64Node) reset() {
	rn.r = nil
	rn.l = nil
	rn.p = nil
	rn.v = u64Empty
}

func (rn *radixU64Node) setL(l *radixU64Node) {
	rn.l = l
}

func (rn *radixU64Node) setR(r *radixU64Node) {
	rn.r = r
}

func (rn *radixU64Node) setP(p *radixU64Node) {
	rn.p = p
}

func (rn *radixU64Node) setV(v uint64) {
	rn.v = v
}

func (rn *radixU64Node) getL() *radixU64Node {
	return rn.l
}

func (rn *radixU64Node) getR() *radixU64Node {
	return rn.r
}

func (rn *radixU64Node) getP() *radixU64Node {
	return rn.p
}

func (rn *radixU64Node) getV() uint64 {
	return rn.v
}

type radixTree struct {
	u32r    *radixU32Node //根节点
	u64r    *radixU64Node //根节点
	u32Pool *sync.Pool
	u64Pool *sync.Pool
}

type IRadixTree interface {
	AddU32(key uint32, mask uint32, value uint32) error
	DelU32(key uint32, mask uint32) error
	GetU32(key uint32) (uint32, error)
	ListU32() []uint32

	AddU64(key uint64, mask uint64, value uint64) error
	DelU64(key uint64, mask uint64) error
	GetU64(key uint64) (uint64, error)
	ListU64() []uint64
}

func NewRadixTree() IRadixTree {
	return newRadixTree()
}

func (t *radixTree) AddU32(key uint32, mask uint32, value uint32) error {
	return t.addU32(key, mask, value)
}

func (t *radixTree) DelU32(key uint32, mask uint32) error {
	return t.delU32(key, mask)
}

func (t *radixTree) GetU32(key uint32) (uint32, error) {
	return t.getU32(key)
}

func (t *radixTree) ListU32() []uint32 {
	return t.listU32()
}

func (t *radixTree) AddU64(key uint64, mask uint64, value uint64) error {
	return t.addU64(key, mask, value)
}

func (t *radixTree) DelU64(key uint64, mask uint64) error {
	return t.delU64(key, mask)
}

func (t *radixTree) GetU64(key uint64) (uint64, error) {
	return t.getU64(key)
}

func (t *radixTree) ListU64() []uint64 {
	return t.listU64()
}

func newRadixTree() *radixTree {
	var tree = new(radixTree)

	tree.u32Pool = &sync.Pool{
		New: func() interface{} {
			return newRadixU32Node()
		},
	}
	tree.u64Pool = &sync.Pool{
		New: func() interface{} {
			return newRadixU64Node()
		},
	}
	tree.u32r = tree.u32Pool.Get().(*radixU32Node)
	tree.u64r = tree.u64Pool.Get().(*radixU64Node)

	return tree
}

func (t *radixTree) addU32(key uint32, mask uint32, value uint32) error {
	var (
		p   = t.u32r //父节点
		c   = p      //子节点
		bit = uint32(0x80000000)
	)

	for bit&mask > 0 {
		if key&bit > 0 {
			c = p.getR()
		} else {
			c = p.getL()
		}
		if nil == c {
			break
		}
		bit >>= 1
		p = c
	}

	//eg:
	//step1: 10.10.10.11/32
	//step2: 10.10.10.12/24
	if c != nil { //掩码影响
		c.setV(value)
	} else {
		for bit&mask > 0 {
			c = t.u32Pool.Get().(*radixU32Node)
			c.reset() //申请节点需要重置
			c.setP(p) //设置父节点

			if bit&key > 0 {
				p.setR(c)
			} else {
				p.setL(c)
			}

			bit >>= 1
			p = c
		}
		c.setV(value)
	}
	return nil
}

func (t *radixTree) delU32(key uint32, mask uint32) error {
	var (
		p   = t.u32r //父节点
		c   = p      //子节点
		bit = uint32(0x80000000)
	)
	for c != nil && bit&mask > 0 {
		if bit&key > 0 {
			c = c.getR()
		} else {
			c = c.getL()
		}
		bit >>= 1
	}
	if nil == c {
		return nil //not found
	}

	if c.isLeaf() { //叶节点
		if c.getP().getL() == c {
			c.getP().setL(nil)
		}
		if c.getP().getR() == c {
			c.getP().setR(nil)
		}
		//回收节点
		t.u32Pool.Put(c)
	} else { //非叶节点
		c.setV(u32Empty)
	}
	return nil
}

func (t *radixTree) getU32(key uint32) (uint32, error) {
	var (
		p   = t.u32r //父节点
		c   = p      //子节点
		v   = u32Empty
		err = notFound
		bit = uint32(0x80000000)
	)
	for c != nil {
		if !c.isEmpty() {
			v = c.getV()
			err = nil
		}
		if bit&key > 0 {
			c = c.getR()
		} else {
			c = c.getL()
		}
		bit >>= 1
	}
	return v, err
}

func (t *radixTree) listU32() []uint32 {
	//中序遍历
	var (
		values []uint32
		p      = t.u32r //父节点
		stack  = make([]*radixU32Node, 32)
		pc     = 0
	)

	for pc > 0 || p != nil {

		for p != nil {
			stack[pc] = p
			pc++
			p = p.getL()
		}

		p = stack[pc-1]
		pc--

		if !p.isEmpty() {
			values = append(values, p.getV())
		}

		p = p.getR()

	}

	return values
}

func (t *radixTree) addU64(key uint64, mask uint64, value uint64) error {
	var (
		p   = t.u64r //父节点
		c   = p      //子节点
		bit = uint64(0x8000000000000000)
	)

	for bit&mask > 0 {
		if key&bit > 0 {
			c = p.getR()
		} else {
			c = p.getL()
		}
		if nil == c {
			break
		}
		bit >>= 1
		p = c
	}

	if c != nil { //掩码影响
		c.setV(value)
	} else {
		for bit&mask > 0 {
			c = t.u64Pool.Get().(*radixU64Node)
			c.reset() //申请节点需要重置
			c.setP(p) //设置父节点

			if bit&key > 0 {
				p.setR(c)
			} else {
				p.setL(c)
			}

			bit >>= 1
			p = c
		}
		c.setV(value)
	}
	return nil
}

func (t *radixTree) delU64(key uint64, mask uint64) error {
	var (
		p   = t.u64r //父节点
		c   = p      //子节点
		bit = uint64(0x8000000000000000)
	)
	for c != nil && bit&mask > 0 {
		if bit&key > 0 {
			c = c.getR()
		} else {
			c = c.getL()
		}
		bit >>= 1
	}
	if nil == c {
		return nil //not found
	}

	if c.isLeaf() { //叶节点
		if c.getP().getL() == c {
			c.getP().setL(nil)
		}
		if c.getP().getR() == c {
			c.getP().setR(nil)
		}
		//回收节点
		t.u64Pool.Put(c)
	} else { //非叶节点
		c.setV(u64Empty)
	}
	return nil
}

func (t *radixTree) getU64(key uint64) (uint64, error) {
	var (
		p   = t.u64r //父节点
		c   = p      //子节点
		v   = u64Empty
		err = notFound
		bit = uint64(0x8000000000000000)
	)
	for c != nil {
		if !c.isEmpty() {
			v = c.getV()
			err = nil
		}
		if bit&key > 0 {
			c = c.getR()
		} else {
			c = c.getL()
		}
		bit >>= 1
	}
	return v, err
}

func (t *radixTree) listU64() []uint64 {
	//中序遍历
	var (
		values []uint64
		p      = t.u64r //父节点
		stack  = make([]*radixU64Node, 64)
		pc     = 0
	)

	for pc > 0 || p != nil {

		for p != nil {
			stack[pc] = p
			pc++
			p = p.getL()
		}

		p = stack[pc-1]
		pc--

		if !p.isEmpty() {
			values = append(values, p.getV())
		}

		p = p.getR()

	}

	return values
}
