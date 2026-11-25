// Copyright 2014 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// 来源：https://github.com/google/btree/blob/479b5e81b0a93ec038d201b0b33d17db599531d3/btree.go
// 本地修改：拆分文件，适配 3rd 接口和 Option 构造；保留原节点算法。
package btree

// BTree is an implementation of a B-Tree.
//
// BTree stores Item instances in an ordered structure, allowing easy insertion,
// removal, and iteration.
//
// Write operations are not safe for concurrent mutation by multiple
// goroutines, but Read operations are.
type btree struct {
	degree int
	length int
	root   *node
	cow    *copyOnWriteContext
}

// copyOnWriteContext pointers determine node ownership... a tree with a write
// context equivalent to a node's write context is allowed to modify that node.
// A tree whose write context does not match a node's is not allowed to modify
// it, and must create a new, writable copy (IE: it's a Clone).
//
// When doing any write operation, we maintain the invariant that the current
// node's context is equal to the context of the tree that requested the write.
// We do this by, before we descend into any node, creating a copy with the
// correct context if the contexts don't match.
//
// Since the node we're currently visiting on any write has the requesting
// tree's context, that node is modifiable in place.  Children of that node may
// not share context, but before we descend into them, we'll make a mutable
// copy.
type copyOnWriteContext struct {
	freelist *FreeList
}

// Clone clones the btree, lazily.  Clone should not be called concurrently,
// but the original tree (t) and the new tree (t2) can be used concurrently
// once the Clone call completes.
//
// The internal tree structure of b is marked read-only and shared between t and
// t2.  Writes to both t and t2 use copy-on-write logic, creating new nodes
// whenever one of b's original nodes would have been modified.  Read operations
// should have no performance degredation.  Write operations for both t and t2
// will initially experience minor slow-downs caused by additional allocs and
// copies due to the aforementioned copy-on-write logic, but should converge to
// the original performance characteristics of the original tree.
func (t *btree) Clone() (t2 BTree) {
	// Create two entirely new copy-on-write contexts.
	// This operation effectively creates three trees:
	//   the original, shared nodes (old b.cow)
	//   the new b.cow nodes
	//   the new out.cow nodes
	cow1, cow2 := *t.cow, *t.cow
	out := *t
	t.cow = &cow1
	out.cow = &cow2
	return &out
}

// maxItems returns the max number of items to allow per node.
func (t *btree) maxItems() int {
	return t.degree*2 - 1
}

// minItems returns the min number of items to allow per node (ignored for the
// root node).
func (t *btree) minItems() int {
	return t.degree - 1
}

func (c *copyOnWriteContext) newNode() (n *node) {
	n = c.freelist.newNode()
	n.cow = c
	return
}

type freeType int

const (
	ftFreelistFull freeType = iota // node was freed (available for GC, not stored in freelist)
	ftStored                       // node was stored in the freelist for later use
	ftNotOwned                     // node was ignored by COW, since it's owned by another one
)

// freeNode frees a node within a given COW context, if it's owned by that
// context.  It returns what happened to the node (see freeType const
// documentation).
func (c *copyOnWriteContext) freeNode(n *node) freeType {
	if n.cow == c {
		// clear to allow GC
		n.items.truncate(0)
		n.children.truncate(0)
		n.cow = nil
		if c.freelist.freeNode(n) {
			return ftStored
		} else {
			return ftFreelistFull
		}
	} else {
		return ftNotOwned
	}
}

// ReplaceOrInsert adds the given item to the tree.  If an item in the tree
// already equals the given one, it is removed from the tree and returned.
// Otherwise, nil is returned.
//
// nil cannot be added to the tree (will panic).
func (t *btree) ReplaceOrInsert(item Item) Item {
	if item == nil {
		panic("nil item being added to BTree")
	}
	if t.root == nil {
		t.root = t.cow.newNode()
		t.root.items = append(t.root.items, item)
		t.length++
		return nil
	} else {
		t.root = t.root.mutableFor(t.cow)
		if len(t.root.items) >= t.maxItems() {
			item2, second := t.root.split(t.maxItems() / 2)
			oldroot := t.root
			t.root = t.cow.newNode()
			t.root.items = append(t.root.items, item2)
			t.root.children = append(t.root.children, oldroot, second)
		}
	}
	out := t.root.insert(item, t.maxItems())
	if out == nil {
		t.length++
	}
	return out
}

// Delete removes an item equal to the passed in item from the tree, returning
// it.  If no such item exists, returns nil.
func (t *btree) Delete(item Item) Item {
	return t.deleteItem(item, removeItem)
}

// DeleteMin removes the smallest item in the tree and returns it.
// If no such item exists, returns nil.
func (t *btree) DeleteMin() Item {
	return t.deleteItem(nil, removeMin)
}

// DeleteMax removes the largest item in the tree and returns it.
// If no such item exists, returns nil.
func (t *btree) DeleteMax() Item {
	return t.deleteItem(nil, removeMax)
}

func (t *btree) deleteItem(item Item, typ toRemove) Item {
	if t.root == nil || len(t.root.items) == 0 {
		return nil
	}
	t.root = t.root.mutableFor(t.cow)
	out := t.root.remove(item, t.minItems(), typ)
	if len(t.root.items) == 0 && len(t.root.children) > 0 {
		oldroot := t.root
		t.root = t.root.children[0]
		t.cow.freeNode(oldroot)
	}
	if out != nil {
		t.length--
	}
	return out
}

// AscendRange calls the iterator for every value in the tree within the range
// [greaterOrEqual, lessThan), until iterator returns false.
func (t *btree) AscendRange(greaterOrEqual, lessThan Item, iterator ItemIterator) {
	if t.root == nil {
		return
	}
	t.root.iterate(ascend, greaterOrEqual, lessThan, true, false, iterator)
}

// AscendLessThan calls the iterator for every value in the tree within the range
// [first, pivot), until iterator returns false.
func (t *btree) AscendLessThan(pivot Item, iterator ItemIterator) {
	if t.root == nil {
		return
	}
	t.root.iterate(ascend, nil, pivot, false, false, iterator)
}

// AscendGreaterOrEqual calls the iterator for every value in the tree within
// the range [pivot, last], until iterator returns false.
func (t *btree) AscendGreaterOrEqual(pivot Item, iterator ItemIterator) {
	if t.root == nil {
		return
	}
	t.root.iterate(ascend, pivot, nil, true, false, iterator)
}

// Ascend calls the iterator for every value in the tree within the range
// [first, last], until iterator returns false.
func (t *btree) Ascend(iterator ItemIterator) {
	if t.root == nil {
		return
	}
	t.root.iterate(ascend, nil, nil, false, false, iterator)
}

// DescendRange calls the iterator for every value in the tree within the range
// [lessOrEqual, greaterThan), until iterator returns false.
func (t *btree) DescendRange(lessOrEqual, greaterThan Item, iterator ItemIterator) {
	if t.root == nil {
		return
	}
	t.root.iterate(descend, lessOrEqual, greaterThan, true, false, iterator)
}

// DescendLessOrEqual calls the iterator for every value in the tree within the range
// [pivot, first], until iterator returns false.
func (t *btree) DescendLessOrEqual(pivot Item, iterator ItemIterator) {
	if t.root == nil {
		return
	}
	t.root.iterate(descend, pivot, nil, true, false, iterator)
}

// DescendGreaterThan calls the iterator for every value in the tree within
// the range [last, pivot), until iterator returns false.
func (t *btree) DescendGreaterThan(pivot Item, iterator ItemIterator) {
	if t.root == nil {
		return
	}
	t.root.iterate(descend, nil, pivot, false, false, iterator)
}

// Descend calls the iterator for every value in the tree within the range
// [last, first], until iterator returns false.
func (t *btree) Descend(iterator ItemIterator) {
	if t.root == nil {
		return
	}
	t.root.iterate(descend, nil, nil, false, false, iterator)
}

// Get looks for the key item in the tree, returning it.  It returns nil if
// unable to find that item.
func (t *btree) Get(key Item) Item {
	if t.root == nil {
		return nil
	}
	return t.root.get(key)
}

// Min returns the smallest item in the tree, or nil if the tree is empty.
func (t *btree) Min() Item {
	return min(t.root)
}

// Max returns the largest item in the tree, or nil if the tree is empty.
func (t *btree) Max() Item {
	return max(t.root)
}

// Has returns true if the given key is in the tree.
func (t *btree) Has(key Item) bool {
	return t.Get(key) != nil
}

// Len returns the number of items currently in the tree.
func (t *btree) Len() int {
	return t.length
}

// Clear removes all items from the btree.  If addNodesToFreelist is true,
// t's nodes are added to its freelist as part of this call, until the freelist
// is full.  Otherwise, the root node is simply dereferenced and the subtree
// left to Go's normal GC processes.
//
// This can be much faster
// than calling Delete on all elements, because that requires finding/removing
// each element in the tree and updating the tree accordingly.  It also is
// somewhat faster than creating a new tree to replace the old one, because
// nodes from the old tree are reclaimed into the freelist for use by the new
// one, instead of being lost to the garbage collector.
//
// This call takes:
//
//	O(1): when addNodesToFreelist is false, this is a single operation.
//	O(1): when the freelist is already full, it breaks out immediately
//	O(freelist size):  when the freelist is empty and the nodes are all owned
//	    by this tree, nodes are added to the freelist until full.
//	O(tree size):  when all nodes are owned by another tree, all nodes are
//	    iterated over looking for nodes to add to the freelist, and due to
//	    ownership, none are.
func (t *btree) Clear(addNodesToFreelist bool) {
	if t.root != nil && addNodesToFreelist {
		t.root.reset(t.cow)
	}
	t.root, t.length = nil, 0
}

// reset returns a subtree to the freelist.  It breaks out immediately if the
// freelist is full, since the only benefit of iterating is to fill that
// freelist up.  Returns true if parent reset call should continue.
func (n *node) reset(c *copyOnWriteContext) bool {
	for _, child := range n.children {
		if !child.reset(c) {
			return false
		}
	}
	return c.freeNode(n) != ftFreelistFull
}
