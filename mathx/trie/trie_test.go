package trie

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_trie(t *testing.T) {
	data := []struct {
		text   string   // 待匹配文本
		words  []string // 预置关键字
		remove []string // 匹配前移除的关键字
		found  bool     // 期望是否命中
		word   string   // 期望命中的关键字
	}{
		// 空树、空文本：均不命中。
		{},
		{text: "hello world"},
		{words: []string{"foo"}},
		// 单关键字：命中与未命中。
		{text: "hello world", words: []string{"world"}, found: true, word: "world"},
		{text: "hello world", words: []string{"golang"}},
		// 命中位置在文本中间，验证 start = i+1-depth 的定位。
		{text: "say hi to me", words: []string{"hi"}, found: true, word: "hi"},
		// 关键字即整段文本。
		{text: "abc", words: []string{"abc"}, found: true, word: "abc"},
		// 关键字比文本长：不命中。
		{text: "ab", words: []string{"abc"}},
		// AC 失败指针：ushers 中先经 sh→she 未成尾，回溯命中 he/hers。
		{text: "ushers", words: []string{"he", "she", "his", "hers"}, found: true, word: "she"},
		{text: "this", words: []string{"he", "she", "his", "hers"}, found: true, word: "his"},
		// 公共前缀：ab 与 abc 并存，短词优先命中。
		{text: "xabcy", words: []string{"ab", "abc"}, found: true, word: "ab"},
		{text: "xabcy", words: []string{"abc"}, found: true, word: "abc"},
		// Remove：移除后不再命中。
		{text: "hello world", words: []string{"world"}, remove: []string{"world"}},
		// Remove 公共前缀词不误删更长词：移除 ab 后 abc 仍可命中。
		{text: "xabcy", words: []string{"ab", "abc"}, remove: []string{"ab"}, found: true, word: "abc"},
		// Remove 更长词不影响更短的前缀词。
		{text: "xaby", words: []string{"ab", "abc"}, remove: []string{"abc"}, found: true, word: "ab"},
		// Remove 不存在的词：静默跳过，原词仍命中。
		{text: "hello world", words: []string{"world"}, remove: []string{"golang"}, found: true, word: "world"},
		// 中文 / 多字节 rune。
		{text: "这是敏感词测试", words: []string{"敏感词"}, found: true, word: "敏感词"},
		{text: "这是正常文本", words: []string{"敏感词"}},
	}

	for i, d := range data {
		f := func(t *testing.T) {
			tree, err := NewTrie()
			assert.Nil(t, err)
			assert.NotNil(t, tree)

			tree.Add(d.words...)
			if len(d.remove) > 0 {
				tree.Remove(d.remove...)
			}

			found, word := tree.Match(d.text)
			assert.Equal(t, d.found, found)
			assert.Equal(t, d.word, word)
		}
		t.Run(fmt.Sprintf("case-%d", i), f)
	}
}
