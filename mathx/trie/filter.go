package trie

import (
	"bufio"
	"io"
	"regexp"
)

// filter 高层敏感词过滤器。
type filter struct {
	trie  Trie
	noise *regexp.Regexp
	dirty bool // 是否需要（重新）构建失败指针
}

// NewFilter 新建敏感词过滤器。noisePattern 非法时返回 error。
func NewFilter(opts ...Option) (Filter, error) {
	o := defaultOption
	for _, opt := range opts {
		opt.Apply(&o)
	}
	noise, err := regexp.Compile(o.noisePattern)
	if err != nil {
		return nil, err
	}
	f := &filter{trie: newTrie(), noise: noise}
	if len(o.words) > 0 {
		f.AddWord(o.words...)
	}
	return f, nil
}

// AddWord 批量添加敏感词。
func (f *filter) AddWord(words ...string) {
	f.trie.Add(words...)
	f.dirty = true
}

// AddWordWithValues 批量添加敏感词并各自携带关联值。
func (f *filter) AddWordWithValues(words map[string][]string) {
	for key, value := range words {
		f.trie.AddWithValues(key, value)
	}
	f.dirty = true
}

// Load 从 io.Reader 按行读取敏感词并添加。
func (f *filter) Load(rd io.Reader) error {
	buf := bufio.NewReader(rd)
	for {
		line, _, err := buf.ReadLine()
		if err != nil {
			if err != io.EOF {
				return err
			}
			break
		}
		f.AddWord(string(line))
	}
	return nil
}

// ensureLinks 在首次匹配前（或新增词后）懒构建失败指针。
func (f *filter) ensureLinks() {
	if f.dirty {
		f.trie.BuildFailureLinks()
		f.dirty = false
	}
}

// RemoveNoise 去除空格等噪音字符。
func (f *filter) RemoveNoise(text string) string {
	return f.noise.ReplaceAllString(text, "")
}

// Filter 删除文本中的敏感词。
func (f *filter) Filter(text string) string {
	f.ensureLinks()
	return f.trie.Filter(text)
}

// Replace 将文本中的敏感词替换为 repl。
func (f *filter) Replace(text string, repl rune) string {
	f.ensureLinks()
	return f.trie.Replace(text, repl)
}

// FindIn 检测文本是否含敏感词（先去噪）。
func (f *filter) FindIn(text string) (bool, string) {
	f.ensureLinks()
	return f.trie.FindIn(f.RemoveNoise(text))
}

// FindAll 返回文本中命中的所有敏感词。
func (f *filter) FindAll(text string) []string {
	f.ensureLinks()
	return f.trie.FindAll(text)
}

// Validate 校验文本是否合法（先去噪）。
func (f *filter) Validate(text string) (bool, string) {
	f.ensureLinks()
	return f.trie.Validate(f.RemoveNoise(text))
}

// MatchFirst 返回第一个命中的敏感词。
func (f *filter) MatchFirst(text string, opts ...MatchOption) (bool, string) {
	f.ensureLinks()
	o := defaultMatchOption
	for _, opt := range opts {
		opt.Apply(&o)
	}
	if o.removeNoise {
		text = f.RemoveNoise(text)
	}
	return f.trie.MatchFirst(text)
}

// MatchFirstAndValues 返回第一个命中的敏感词及其关联值。
func (f *filter) MatchFirstAndValues(text string, opts ...MatchOption) (bool, string, []string) {
	f.ensureLinks()
	o := defaultMatchOption
	for _, opt := range opts {
		opt.Apply(&o)
	}
	if o.removeNoise {
		text = f.RemoveNoise(text)
	}
	return f.trie.MatchFirstAndValues(text)
}
