package btree

import (
	"fmt"
	"math/rand"
	"reflect"
	"sort"
	"sync"
	"testing"
)

func TestConfiguration(t *testing.T) {
	var nilOption *structOption
	for _, opts := range [][]Option{{nil}, {nilOption}, {WithDegree(-1)}, {WithDegree(0)}, {WithDegree(1)}, {WithDegree(int(^uint(0) >> 1))}, {WithFreeList(nil)}} {
		if tree, err := NewBTree(opts...); err == nil || tree != nil {
			t.Fatalf("invalid configuration accepted: %v", opts)
		}
	}
	if f, err := NewFreeList(-1); err == nil || f != nil {
		t.Fatal("negative capacity accepted")
	}
	zero := mustFreeList(t, 0)
	tree, err := NewBTree(WithDegree(1), WithDegree(2), WithFreeList(zero))
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range perm(100) {
		tree.ReplaceOrInsert(v)
	}
	tree.Clear(true)
	if tree.Len() != 0 {
		t.Fatal("clear failed with zero cache capacity")
	}
	a, err := NewBTree()
	if err != nil {
		t.Fatal(err)
	}
	b, err := NewBTree()
	if err != nil {
		t.Fatal(err)
	}
	if a.(*btree).cow.freelist == b.(*btree).cow.freelist {
		t.Fatal("default cache is shared")
	}
}

type structOption struct{}

func (*structOption) Apply(*option) {}

func TestEmptyAndEarlyStop(t *testing.T) {
	tree := mustTree(t, 2)
	if tree.Len() != 0 || tree.Get(Int(1)) != nil || tree.Has(Int(1)) || tree.Delete(Int(1)) != nil || tree.DeleteMin() != nil || tree.DeleteMax() != nil {
		t.Fatal("incorrect empty-tree result")
	}
	tree.Ascend(func(Item) bool { t.Fatal("empty tree visited"); return false })
	tree.Descend(func(Item) bool { t.Fatal("empty tree visited"); return false })
	for _, v := range perm(100) {
		tree.ReplaceOrInsert(v)
	}
	for _, iterate := range []func(ItemIterator){tree.Ascend, tree.Descend} {
		count := 0
		iterate(func(Item) bool { count++; return false })
		if count != 1 {
			t.Fatalf("visited %d items", count)
		}
	}
	defer func() {
		if recover() == nil {
			t.Fatal("nil insertion did not panic")
		}
	}()
	tree.ReplaceOrInsert(nil)
}

func TestClearCloneIsolation(t *testing.T) {
	for _, recycle := range []bool{false, true} {
		t.Run(fmt.Sprint(recycle), func(t *testing.T) {
			tree := mustTree(t, 2)
			for _, v := range perm(200) {
				tree.ReplaceOrInsert(v)
			}
			clone := tree.Clone()
			tree.Clear(recycle)
			for i := 200; i < 400; i++ {
				tree.ReplaceOrInsert(Int(i))
			}
			if got := all(clone); !reflect.DeepEqual(got, rang(200)) {
				t.Fatal("clear or cache reuse changed clone")
			}
			clone.Delete(Int(1))
			clone.ReplaceOrInsert(Int(-1))
			if tree.Len() != 200 || tree.Min() != Int(200) || tree.Max() != Int(399) {
				t.Fatal("clone mutation changed source")
			}
		})
	}
}

// Compare mixed operations with a map and sorted keys, including minimum degree.
func TestMixedOperations(t *testing.T) {
	for _, degree := range []int{2, 3, 32} {
		t.Run(fmt.Sprint(degree), func(t *testing.T) {
			tree := mustTree(t, degree)
			want := map[int]bool{}
			rng := rand.New(rand.NewSource(42))
			for i := 0; i < 4000; i++ {
				key := rng.Intn(200)
				exists := want[key]
				switch rng.Intn(4) {
				case 0, 1:
					old := tree.ReplaceOrInsert(Int(key))
					if (old != nil) != exists || old != nil && old != Int(key) {
						t.Fatal("replace result mismatch")
					}
					want[key] = true
				case 2:
					old := tree.Delete(Int(key))
					if (old != nil) != exists || old != nil && old != Int(key) {
						t.Fatal("delete result mismatch")
					}
					delete(want, key)
				case 3:
					if tree.Has(Int(key)) != exists {
						t.Fatal("has mismatch")
					}
				}
				keys := make([]int, 0, len(want))
				for k := range want {
					keys = append(keys, k)
				}
				sort.Ints(keys)
				got := all(tree)
				if len(got) != len(keys) || tree.Len() != len(keys) {
					t.Fatal("length mismatch")
				}
				for j, k := range keys {
					if got[j] != Int(k) {
						t.Fatal("order mismatch")
					}
				}
			}
			for tree.Len() > 0 {
				tree.DeleteMin()
			}
			if tree.Min() != nil || tree.Max() != nil {
				t.Fatal("root did not shrink to empty")
			}
			tree.ReplaceOrInsert(Int(7))
			if tree.Get(Int(7)) != Int(7) {
				t.Fatal("reuse after empty failed")
			}
		})
	}
}

func TestSharedFreeList(t *testing.T) {
	f := mustFreeList(t, 32)
	trees := make([]BTree, 4)
	for i := range trees {
		trees[i] = mustTree(t, 2, WithFreeList(f))
	}
	var wg sync.WaitGroup
	for _, tree := range trees {
		wg.Add(1)
		go func(tree BTree) {
			defer wg.Done()
			for pass := 0; pass < 10; pass++ {
				for i := 0; i < 100; i++ {
					tree.ReplaceOrInsert(Int(i))
				}
				tree.Clear(true)
			}
			tree.ReplaceOrInsert(Int(42))
		}(tree)
	}
	wg.Wait()
	for _, tree := range trees {
		if tree.Len() != 1 || tree.Get(Int(42)) != Int(42) {
			t.Fatal("shared cache corrupted tree")
		}
	}
}
