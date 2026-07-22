package trie

import "regexp"

// filter 高层敏感词过滤器。
type filter struct {
	trie  *trie
	noise *regexp.Regexp
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
}

// RemoveWord 批量移除敏感词；不存在的词静默跳过。
func (f *filter) RemoveWord(words ...string) {
	f.trie.Remove(words...)
}

// removeNoise 去除空格等噪音字符。
func (f *filter) removeNoise(text string) string {
	return f.noise.ReplaceAllString(text, "")
}

// Match 返回命中的敏感词，默认匹配第一个。
func (f *filter) Match(text string, opts ...MatchOption) (bool, string) {
	o := defaultMatchOption
	for _, opt := range opts {
		opt.Apply(&o)
	}
	if o.removeNoise {
		text = f.removeNoise(text)
	}
	return f.trie.Match(text)
}
