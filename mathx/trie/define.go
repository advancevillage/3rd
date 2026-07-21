// Package trie 提供基于 Trie 树 + Aho-Corasick 自动机的关键字（敏感词）匹配能力。
//
// 该包仅使用 Go 标准库与本模块内的 x 包，不依赖任何第三方库。
// 对外暴露两层接口：
//
//   - Trie   底层 AC 自动机，负责关键字的增加与各种匹配算法。
//   - Filter 高层敏感词过滤器，在 Trie 之上提供内存加载、去噪、失败指针懒构建。
package trie

import "io"

// Trie 关键字匹配树（底层 Aho-Corasick 自动机）。
//
// 使用流程：先 Add 添加关键字，再调用 BuildFailureLinks 构建失败指针，
// 之后即可进行各种匹配。
type Trie interface {
	// Add 批量添加关键字。
	Add(words ...string)
	// AddWithValues 添加一个关键字并关联一组自定义值。
	AddWithValues(word string, values []string)
	// BuildFailureLinks 构建 AC 失败指针；Add 结束后、匹配前需调用一次。
	BuildFailureLinks()

	// Replace 将文本中命中的关键字替换为 character。
	Replace(text string, character rune) string
	// Filter 直接删除文本中命中的关键字。
	Filter(text string) string
	// Validate 校验文本是否合法；不合法时返回 false 及命中的第一个关键字。
	Validate(text string) (ok bool, first string)
	// FindIn 判断文本是否含关键字；命中时返回 true 及命中的第一个关键字。
	FindIn(text string) (found bool, first string)
	// FindAll 返回文本中命中的所有关键字。
	FindAll(text string) []string
	// MatchFirst 返回第一个命中的关键字，未命中返回 false 与空串。
	MatchFirst(text string) (found bool, word string)
	// MatchFirstAndValues 返回第一个命中的关键字及其关联值。
	MatchFirstAndValues(text string) (found bool, word string, values []string)
	// BatchMatchFirst 对多段文本依次匹配，返回首个命中。
	BatchMatchFirst(texts []string) (found bool, word string)
	// BatchMatchFirstAndValues 对多段文本依次匹配，返回首个命中及其关联值。
	BatchMatchFirstAndValues(texts []string) (found bool, word string, values []string)
}

// Filter 高层敏感词过滤器：内存加载关键字 + 去噪 + 失败指针懒构建。
type Filter interface {
	// AddWord 批量添加敏感词。
	AddWord(words ...string)
	// AddWordWithValues 批量添加敏感词并各自携带关联值。
	AddWordWithValues(words map[string][]string)
	// Load 从 io.Reader 按行读取敏感词并添加。
	Load(rd io.Reader) error

	// Filter 删除文本中的敏感词。
	Filter(text string) string
	// Replace 将文本中的敏感词替换为 repl。
	Replace(text string, repl rune) string
	// Validate 校验文本是否合法（会先去噪）。
	Validate(text string) (ok bool, first string)
	// FindIn 检测文本是否含敏感词（会先去噪）。
	FindIn(text string) (found bool, first string)
	// FindAll 返回文本中命中的所有敏感词。
	FindAll(text string) []string
	// MatchFirst 返回第一个命中的敏感词；可用 WithRemoveNoise 控制匹配前是否先去噪。
	MatchFirst(text string, opts ...MatchOption) (found bool, word string)
	// MatchFirstAndValues 返回第一个命中的敏感词及其关联值。
	MatchFirstAndValues(text string, opts ...MatchOption) (found bool, word string, values []string)
	// RemoveNoise 去除文本中的噪音字符（空格、特殊符号等）。
	RemoveNoise(text string) string
}

var (
	_ Trie   = (*trie)(nil)
	_ Filter = (*filter)(nil)
)
