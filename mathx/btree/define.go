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

// Item represents a single object in the tree.
type Item interface {
	// Less tests whether the current item is less than the given argument.
	//
	// This must provide a strict weak ordering.
	// If !a.Less(b) && !b.Less(a), we treat this to mean a == b (i.e. we can only
	// hold one of either a or b in the tree).
	Less(than Item) bool
}

// ItemIterator allows callers of Ascend* to iterate in-order over portions of
// the tree.  When this function returns false, iteration will stop and the
// associated Ascend* function will immediately return.
type ItemIterator func(i Item) bool

// BTree 有序内存树，由 NewBTree 构造；不支持直接并发读写。
// Clone 共享 Item；Clear 的参数决定是否将节点归还缓存。
type BTree interface {
	// ReplaceOrInsert adds the given item to the tree.  If an item in the tree
	// already equals the given one, it is removed from the tree and returned.
	// Otherwise, nil is returned.
	//
	// nil cannot be added to the tree (will panic).
	ReplaceOrInsert(Item) Item
	// Delete removes an item equal to the passed in item from the tree, returning
	// it.  If no such item exists, returns nil.
	Delete(Item) Item
	// DeleteMin removes the smallest item in the tree and returns it.
	// If no such item exists, returns nil.
	DeleteMin() Item
	// DeleteMax removes the largest item in the tree and returns it.
	// If no such item exists, returns nil.
	DeleteMax() Item
	// Get looks for the key item in the tree, returning it.  It returns nil if
	// unable to find that item.
	Get(Item) Item
	// Has returns true if the given key is in the tree.
	Has(Item) bool
	// Min returns the smallest item in the tree, or nil if the tree is empty.
	Min() Item
	// Max returns the largest item in the tree, or nil if the tree is empty.
	Max() Item
	// Len returns the number of items currently in the tree.
	Len() int
	// Ascend calls the iterator for every value in the tree within the range
	// [first, last], until iterator returns false.
	Ascend(ItemIterator)
	// AscendRange 遍历 [greaterOrEqual, lessThan)。
	// AscendRange calls the iterator for every value in the tree within the range
	// [greaterOrEqual, lessThan), until iterator returns false.
	AscendRange(greaterOrEqual, lessThan Item, iterator ItemIterator)
	// AscendLessThan calls the iterator for every value in the tree within the range
	// [first, pivot), until iterator returns false.
	AscendLessThan(Item, ItemIterator)
	// AscendGreaterOrEqual calls the iterator for every value in the tree within
	// the range [pivot, last], until iterator returns false.
	AscendGreaterOrEqual(Item, ItemIterator)
	// Descend calls the iterator for every value in the tree within the range
	// [last, first], until iterator returns false.
	Descend(ItemIterator)
	// DescendRange 遍历 (greaterThan, lessOrEqual]。
	// DescendRange calls the iterator for every value in the tree within the range
	// [lessOrEqual, greaterThan), until iterator returns false.
	DescendRange(lessOrEqual, greaterThan Item, iterator ItemIterator)
	// DescendLessOrEqual calls the iterator for every value in the tree within the range
	// [pivot, first], until iterator returns false.
	DescendLessOrEqual(Item, ItemIterator)
	// DescendGreaterThan calls the iterator for every value in the tree within
	// the range [last, pivot), until iterator returns false.
	DescendGreaterThan(Item, ItemIterator)
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
	Clone() BTree
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
	//   O(1): when addNodesToFreelist is false, this is a single operation.
	//   O(1): when the freelist is already full, it breaks out immediately
	//   O(freelist size):  when the freelist is empty and the nodes are all owned
	//       by this tree, nodes are added to the freelist until full.
	//   O(tree size):  when all nodes are owned by another tree, all nodes are
	//       iterated over looking for nodes to add to the freelist, and due to
	//       ownership, none are.
	Clear(addNodesToFreelist bool)
}

var _ BTree = (*btree)(nil)

// Int implements the Item interface for integers.
type Int int

// Less returns true if int(a) < int(b).
func (a Int) Less(b Item) bool {
	return a < b.(Int)
}
