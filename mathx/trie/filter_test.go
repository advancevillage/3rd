package trie

import "testing"

func newTestFilter(t *testing.T) Filter {
	t.Helper()
	f, err := NewFilter()
	if err != nil {
		t.Fatalf("new filter: %v", err)
	}
	return f
}

func TestFilterAddRemove(t *testing.T) {
	filter := newTestFilter(t)
	filter.AddWord("一个东西", "一个", "东西")

	if found, word := filter.Match("有一个东西"); !found || word != "一个" {
		t.Fatalf("Match(有一个东西) got %v %s, expect true 一个", found, word)
	}

	// 移除公共前缀词后，共享前缀的其它词仍应命中。
	filter.RemoveWord("一个")
	if found, _ := filter.Match("一个物体"); found {
		t.Fatalf("Match(一个物体) should miss after RemoveWord(一个)")
	}
	if found, word := filter.Match("有一个东西"); !found || word != "一个东西" {
		t.Fatalf("Match(有一个东西) got %v %s, expect true 一个东西", found, word)
	}

	// 移除不存在的词应静默跳过。
	filter.RemoveWord("不存在")
	if found, word := filter.Match("东西"); !found || word != "东西" {
		t.Fatalf("Match(东西) after removing missing word got %v %s", found, word)
	}
}

func TestMatchFirstWithRemoveNoise(t *testing.T) {
	filter := newTestFilter(t)
	filter.AddWord("东西")

	if found, word := filter.Match("有东 西哈", WithRemoveNoise(true)); !found || word != "东西" {
		t.Fatalf("match with noise removal, got %v %s, expect true 东西", found, word)
	}
	if found, _ := filter.Match("有东 西哈"); found {
		t.Fatalf("match without noise removal should miss")
	}
}
