package trie

import (
	"fmt"
	"reflect"
	"testing"
)

func TestTrieTree(t *testing.T) {
	tree := newTrie()

	tree.Add("王麻子", "王麻子")
	tree.BuildFailureLinks()
	fmt.Println(tree.root.children[0])
	fmt.Println(tree.Replace("你好吗 我支持王麻子， 他的名字叫做王麻子", '*'))
	fmt.Println(tree.Filter("你好吗 我支持王麻子， 他的名字叫做王麻子"))
}

// TestTrieTreeBFS 校验 BFS 逐层遍历的结果。
// 注意：同层内的顺序依赖 map 迭代（无序），因此按 depth 归类后比较集合，
// 而不是比较严格顺序。
func TestTrieTreeBFS(t *testing.T) {
	tree := newTrie()
	tree.Add("王麻子", "共产党好")

	expect := map[int]map[string]struct{}{
		1: {"王": {}, "共": {}},
		2: {"麻": {}, "产": {}},
		3: {"子": {}, "党": {}},
		4: {"好": {}},
	}

	got := map[int]map[string]struct{}{}
	for n := range tree.bfs() {
		if got[n.depth] == nil {
			got[n.depth] = map[string]struct{}{}
		}
		got[n.depth][string(n.character)] = struct{}{}
	}

	if !reflect.DeepEqual(expect, got) {
		t.Errorf("bfs mismatch, expect %v got %v", expect, got)
	}
}

func TestTrieTreeBuildFailureLinks(t *testing.T) {
	tree := newTrie()
	tree.Add("he", "his", "she", "hers")
	tree.BuildFailureLinks()
}
