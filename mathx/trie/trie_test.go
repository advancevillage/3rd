package trie

import (
	"reflect"
	"testing"
)

func TestTrieTree(t *testing.T) {
	tree := newTrie()

	tree.Add("王麻子", "王麻子")
	if found, word := tree.Match("你好吗 我支持王麻子， 他的名字叫做王麻子"); !found || word != "王麻子" {
		t.Fatalf("Match got %v %s, expect true 王麻子", found, word)
	}
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
	tree.ensureLinks()
}

func TestTrieRemove(t *testing.T) {
	tree := newTrie()
	tree.Add("一个东西", "一个", "东西")

	// 移除公共前缀词 "一个"，不应影响共享前缀的其它词。
	tree.Remove("一个")

	if found, _ := tree.Match("一个"); found {
		t.Fatalf("Match(一个) should miss after Remove")
	}
	if found, word := tree.Match("一个东西"); !found || word != "一个东西" {
		t.Fatalf("Match(一个东西) got %v %s, expect true 一个东西", found, word)
	}
	if found, word := tree.Match("东西"); !found || word != "东西" {
		t.Fatalf("Match(东西) got %v %s, expect true 东西", found, word)
	}

	// 移除不存在的词应静默跳过，不影响已有词。
	tree.Remove("不存在")
	if found, word := tree.Match("东西"); !found || word != "东西" {
		t.Fatalf("Match(东西) after removing missing word got %v %s", found, word)
	}
}
