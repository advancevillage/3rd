package trie

import "github.com/advancevillage/3rd/x"

// Option 作用于过滤器构造（NewFilter）。
type Option = x.Options[option]

type option struct {
	noisePattern string   // 去噪正则表达式
	words        []string // 预置关键字
}

var defaultOption = option{
	noisePattern: `[\|\s&%$@*]+`,
}

// WithNoisePattern 设置去噪正则（替代旧的 UpdateNoisePattern）。
func WithNoisePattern(pattern string) Option {
	return x.NewFuncOptions(func(o *option) {
		o.noisePattern = pattern
	})
}

// WithWords 在构造时预置一批关键字。
func WithWords(words ...string) Option {
	return x.NewFuncOptions(func(o *option) {
		o.words = append(o.words, words...)
	})
}

// MatchOption 作用于单次匹配（Match）。
type MatchOption = x.Options[matchOption]

type matchOption struct {
	removeNoise bool // 匹配前是否先去噪
}

var defaultMatchOption = matchOption{}

// WithRemoveNoise 控制匹配前是否先去噪。
func WithRemoveNoise(remove bool) MatchOption {
	return x.NewFuncOptions(func(o *matchOption) {
		o.removeNoise = remove
	})
}
