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

// 来源：https://github.com/google/btree/blob/479b5e81b0a93ec038d201b0b33d17db599531d3/btree_test.go
// 本地修改：拆分测试与基准，适配新构造 API，并使用局部固定随机种子。
package btree

import (
	"flag"
	"math/rand"
	"reflect"
	"sync"
	"testing"
)

func mustTree(t testing.TB, degree int, opts ...Option) BTree {
	t.Helper()
	tree, err := NewBTree(append([]Option{WithDegree(degree)}, opts...)...)
	if err != nil {
		t.Fatal(err)
	}
	return tree
}
func mustFreeList(t testing.TB, size int) *FreeList {
	t.Helper()
	f, err := NewFreeList(size)
	if err != nil {
		t.Fatal(err)
	}
	return f
}

// perm returns a random permutation of n Int items in the range [0, n).
func perm(n int) (out []Item) {
	for _, v := range rand.New(rand.NewSource(1)).Perm(n) {
		out = append(out, Int(v))
	}
	return
}

// rang returns an ordered list of Int items in the range [0, n).
func rang(n int) (out []Item) {
	for i := 0; i < n; i++ {
		out = append(out, Int(i))
	}
	return
}

// all extracts all items from a tree in order as a slice.
func all(t BTree) (out []Item) {
	t.Ascend(func(a Item) bool {
		out = append(out, a)
		return true
	})
	return
}

// rangerev returns a reversed ordered list of Int items in the range [0, n).
func rangrev(n int) (out []Item) {
	for i := n - 1; i >= 0; i-- {
		out = append(out, Int(i))
	}
	return
}

// allrev extracts all items from a tree in reverse order as a slice.
func allrev(t BTree) (out []Item) {
	t.Descend(func(a Item) bool {
		out = append(out, a)
		return true
	})
	return
}

var btreeDegree = flag.Int("degree", 32, "B-Tree degree")

func TestBTree(t *testing.T) {
	tr := mustTree(t, *btreeDegree)
	const treeSize = 10000
	for i := 0; i < 10; i++ {
		if min := tr.Min(); min != nil {
			t.Fatalf("empty min, got %+v", min)
		}
		if max := tr.Max(); max != nil {
			t.Fatalf("empty max, got %+v", max)
		}
		for _, item := range perm(treeSize) {
			if x := tr.ReplaceOrInsert(item); x != nil {
				t.Fatal("insert found item", item)
			}
		}
		for _, item := range perm(treeSize) {
			if x := tr.ReplaceOrInsert(item); x == nil {
				t.Fatal("insert didn't find item", item)
			}
		}
		if min, want := tr.Min(), Item(Int(0)); min != want {
			t.Fatalf("min: want %+v, got %+v", want, min)
		}
		if max, want := tr.Max(), Item(Int(treeSize-1)); max != want {
			t.Fatalf("max: want %+v, got %+v", want, max)
		}
		got := all(tr)
		want := rang(treeSize)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("mismatch:\n got: %v\nwant: %v", got, want)
		}

		gotrev := allrev(tr)
		wantrev := rangrev(treeSize)
		if !reflect.DeepEqual(gotrev, wantrev) {
			t.Fatalf("mismatch:\n got: %v\nwant: %v", got, want)
		}

		for _, item := range perm(treeSize) {
			if x := tr.Delete(item); x == nil {
				t.Fatalf("didn't find %v", item)
			}
		}
		if got = all(tr); len(got) > 0 {
			t.Fatalf("some left!: %v", got)
		}
	}
}

func TestDeleteMin(t *testing.T) {
	tr := mustTree(t, 3)
	for _, v := range perm(100) {
		tr.ReplaceOrInsert(v)
	}
	var got []Item
	for v := tr.DeleteMin(); v != nil; v = tr.DeleteMin() {
		got = append(got, v)
	}
	if want := rang(100); !reflect.DeepEqual(got, want) {
		t.Fatalf("ascendrange:\n got: %v\nwant: %v", got, want)
	}
}

func TestDeleteMax(t *testing.T) {
	tr := mustTree(t, 3)
	for _, v := range perm(100) {
		tr.ReplaceOrInsert(v)
	}
	var got []Item
	for v := tr.DeleteMax(); v != nil; v = tr.DeleteMax() {
		got = append(got, v)
	}
	// Reverse our list.
	for i := 0; i < len(got)/2; i++ {
		got[i], got[len(got)-i-1] = got[len(got)-i-1], got[i]
	}
	if want := rang(100); !reflect.DeepEqual(got, want) {
		t.Fatalf("ascendrange:\n got: %v\nwant: %v", got, want)
	}
}

func TestAscendRange(t *testing.T) {
	tr := mustTree(t, 2)
	for _, v := range perm(100) {
		tr.ReplaceOrInsert(v)
	}
	var got []Item
	tr.AscendRange(Int(40), Int(60), func(a Item) bool {
		got = append(got, a)
		return true
	})
	if want := rang(100)[40:60]; !reflect.DeepEqual(got, want) {
		t.Fatalf("ascendrange:\n got: %v\nwant: %v", got, want)
	}
	got = got[:0]
	tr.AscendRange(Int(40), Int(60), func(a Item) bool {
		if a.(Int) > 50 {
			return false
		}
		got = append(got, a)
		return true
	})
	if want := rang(100)[40:51]; !reflect.DeepEqual(got, want) {
		t.Fatalf("ascendrange:\n got: %v\nwant: %v", got, want)
	}
}

func TestDescendRange(t *testing.T) {
	tr := mustTree(t, 2)
	for _, v := range perm(100) {
		tr.ReplaceOrInsert(v)
	}
	var got []Item
	tr.DescendRange(Int(60), Int(40), func(a Item) bool {
		got = append(got, a)
		return true
	})
	if want := rangrev(100)[39:59]; !reflect.DeepEqual(got, want) {
		t.Fatalf("descendrange:\n got: %v\nwant: %v", got, want)
	}
	got = got[:0]
	tr.DescendRange(Int(60), Int(40), func(a Item) bool {
		if a.(Int) < 50 {
			return false
		}
		got = append(got, a)
		return true
	})
	if want := rangrev(100)[39:50]; !reflect.DeepEqual(got, want) {
		t.Fatalf("descendrange:\n got: %v\nwant: %v", got, want)
	}
}
func TestAscendLessThan(t *testing.T) {
	tr := mustTree(t, *btreeDegree)
	for _, v := range perm(100) {
		tr.ReplaceOrInsert(v)
	}
	var got []Item
	tr.AscendLessThan(Int(60), func(a Item) bool {
		got = append(got, a)
		return true
	})
	if want := rang(100)[:60]; !reflect.DeepEqual(got, want) {
		t.Fatalf("ascendrange:\n got: %v\nwant: %v", got, want)
	}
	got = got[:0]
	tr.AscendLessThan(Int(60), func(a Item) bool {
		if a.(Int) > 50 {
			return false
		}
		got = append(got, a)
		return true
	})
	if want := rang(100)[:51]; !reflect.DeepEqual(got, want) {
		t.Fatalf("ascendrange:\n got: %v\nwant: %v", got, want)
	}
}

func TestDescendLessOrEqual(t *testing.T) {
	tr := mustTree(t, *btreeDegree)
	for _, v := range perm(100) {
		tr.ReplaceOrInsert(v)
	}
	var got []Item
	tr.DescendLessOrEqual(Int(40), func(a Item) bool {
		got = append(got, a)
		return true
	})
	if want := rangrev(100)[59:]; !reflect.DeepEqual(got, want) {
		t.Fatalf("descendlessorequal:\n got: %v\nwant: %v", got, want)
	}
	got = got[:0]
	tr.DescendLessOrEqual(Int(60), func(a Item) bool {
		if a.(Int) < 50 {
			return false
		}
		got = append(got, a)
		return true
	})
	if want := rangrev(100)[39:50]; !reflect.DeepEqual(got, want) {
		t.Fatalf("descendlessorequal:\n got: %v\nwant: %v", got, want)
	}
}
func TestAscendGreaterOrEqual(t *testing.T) {
	tr := mustTree(t, *btreeDegree)
	for _, v := range perm(100) {
		tr.ReplaceOrInsert(v)
	}
	var got []Item
	tr.AscendGreaterOrEqual(Int(40), func(a Item) bool {
		got = append(got, a)
		return true
	})
	if want := rang(100)[40:]; !reflect.DeepEqual(got, want) {
		t.Fatalf("ascendrange:\n got: %v\nwant: %v", got, want)
	}
	got = got[:0]
	tr.AscendGreaterOrEqual(Int(40), func(a Item) bool {
		if a.(Int) > 50 {
			return false
		}
		got = append(got, a)
		return true
	})
	if want := rang(100)[40:51]; !reflect.DeepEqual(got, want) {
		t.Fatalf("ascendrange:\n got: %v\nwant: %v", got, want)
	}
}

func TestDescendGreaterThan(t *testing.T) {
	tr := mustTree(t, *btreeDegree)
	for _, v := range perm(100) {
		tr.ReplaceOrInsert(v)
	}
	var got []Item
	tr.DescendGreaterThan(Int(40), func(a Item) bool {
		got = append(got, a)
		return true
	})
	if want := rangrev(100)[:59]; !reflect.DeepEqual(got, want) {
		t.Fatalf("descendgreaterthan:\n got: %v\nwant: %v", got, want)
	}
	got = got[:0]
	tr.DescendGreaterThan(Int(40), func(a Item) bool {
		if a.(Int) < 50 {
			return false
		}
		got = append(got, a)
		return true
	})
	if want := rangrev(100)[:50]; !reflect.DeepEqual(got, want) {
		t.Fatalf("descendgreaterthan:\n got: %v\nwant: %v", got, want)
	}
}

type byInts []Item

func (a byInts) Len() int {
	return len(a)
}

func (a byInts) Less(i, j int) bool {
	return a[i].(Int) < a[j].(Int)
}

func (a byInts) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

const cloneTestSize = 10000

func cloneTest(t *testing.T, b BTree, start int, p []Item, wg *sync.WaitGroup, trees *[]BTree, lock *sync.Mutex) {
	t.Logf("Starting new clone at %v", start)
	lock.Lock()
	*trees = append(*trees, b)
	lock.Unlock()
	for i := start; i < cloneTestSize; i++ {
		b.ReplaceOrInsert(p[i])
		if i%(cloneTestSize/5) == 0 {
			wg.Add(1)
			go cloneTest(t, b.Clone(), i+1, p, wg, trees, lock)
		}
	}
	wg.Done()
}

func TestCloneConcurrentOperations(t *testing.T) {
	b := mustTree(t, *btreeDegree)
	trees := []BTree{}
	p := perm(cloneTestSize)
	var wg sync.WaitGroup
	wg.Add(1)
	go cloneTest(t, b, 0, p, &wg, &trees, &sync.Mutex{})
	wg.Wait()
	want := rang(cloneTestSize)
	t.Logf("Starting equality checks on %d trees", len(trees))
	for i, tree := range trees {
		if !reflect.DeepEqual(want, all(tree)) {
			t.Errorf("tree %v mismatch", i)
		}
	}
	t.Log("Removing half from first half")
	toRemove := rang(cloneTestSize)[cloneTestSize/2:]
	for i := 0; i < len(trees)/2; i++ {
		tree := trees[i]
		wg.Add(1)
		go func() {
			for _, item := range toRemove {
				tree.Delete(item)
			}
			wg.Done()
		}()
	}
	wg.Wait()
	t.Log("Checking all values again")
	for i, tree := range trees {
		var wantpart []Item
		if i < len(trees)/2 {
			wantpart = want[:cloneTestSize/2]
		} else {
			wantpart = want
		}
		if got := all(tree); !reflect.DeepEqual(wantpart, got) {
			t.Errorf("tree %v mismatch, want %v got %v", i, len(want), len(got))
		}
	}
}
