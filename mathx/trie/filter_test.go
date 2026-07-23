package trie

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_filter(t *testing.T) {
	data := []struct {
		opts        []Option      // 构造选项
		addWords    []string      // 构造后追加的关键字
		removeWords []string      // 匹配前移除的关键字
		text        string        // 待匹配文本
		matchOpts   []MatchOption // 单次匹配选项
		wantErr     bool          // 构造是否期望报错
		found       bool
		word        string
	}{
		// 空过滤器、空文本：均不命中。
		{},
		{text: "hello world"},
		{opts: []Option{WithWords("foo")}},
		// WithWords 预置关键字命中。
		{opts: []Option{WithWords("world")}, text: "hello world", found: true, word: "world"},
		// AddWord 追加关键字命中。
		{addWords: []string{"world"}, text: "hello world", found: true, word: "world"},
		// RemoveWord 后不再命中。
		{opts: []Option{WithWords("world")}, removeWords: []string{"world"}, text: "hello world"},
		// 默认不去噪：噪音字符隔断关键字，不命中。
		{opts: []Option{WithWords("hello")}, text: "he llo"},
		// WithRemoveNoise：去噪后命中。
		{opts: []Option{WithWords("hello")}, text: "he llo", matchOpts: []MatchOption{WithRemoveNoise(true)}, found: true, word: "hello"},
		// 默认噪音模式覆盖多种噪音字符。
		{opts: []Option{WithWords("abc")}, text: "a|b&c", matchOpts: []MatchOption{WithRemoveNoise(true)}, found: true, word: "abc"},
		// 自定义噪音模式。
		{opts: []Option{WithWords("abc"), WithNoisePattern(`[-]+`)}, text: "a-b-c", matchOpts: []MatchOption{WithRemoveNoise(true)}, found: true, word: "abc"},
		// 自定义噪音模式不影响默认噪音字符（空格不被去除）。
		{opts: []Option{WithWords("hello"), WithNoisePattern(`[-]+`)}, text: "he llo", matchOpts: []MatchOption{WithRemoveNoise(true)}},
		// 非法正则：构造报错。
		{opts: []Option{WithNoisePattern(`[invalid`)}, wantErr: true},
		// 中文关键字。
		{opts: []Option{WithWords("敏感词")}, text: "这是敏感词测试", found: true, word: "敏感词"},
		// 中文关键字 + 去噪。
		{opts: []Option{WithWords("敏感词")}, text: "敏 感 词", matchOpts: []MatchOption{WithRemoveNoise(true)}, found: true, word: "敏感词"},
	}

	for i, d := range data {
		f := func(t *testing.T) {
			filter, err := NewFilter(d.opts...)
			if d.wantErr {
				assert.NotNil(t, err)
				return
			}
			assert.Nil(t, err)
			assert.NotNil(t, filter)

			filter.AddWord(d.addWords...)
			if len(d.removeWords) > 0 {
				filter.RemoveWord(d.removeWords...)
			}

			found, word := filter.Match(d.text, d.matchOpts...)
			assert.Equal(t, d.found, found)
			assert.Equal(t, d.word, word)
		}
		t.Run(fmt.Sprintf("case-%d", i), f)
	}
}
