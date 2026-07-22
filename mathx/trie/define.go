// Package trie 提供基于 Trie 树 + Aho-Corasick 自动机的关键字（敏感词）匹配能力。
//
// 该包仅使用 Go 标准库与本模块内的 x 包，不依赖任何第三方库。
// 对外暴露两层接口：
//
//   - Trie   底层 AC 自动机，负责关键字的增删与匹配。
//   - Filter 高层敏感词过滤器，在 Trie 之上提供去噪等能力。
package trie

// Trie 关键字匹配树（底层 Aho-Corasick 自动机）。
//
// 使用流程：Add 添加关键字后即可直接匹配，失败指针在首次匹配时内部懒构建。
type Trie interface {
	// Add 批量添加关键字。
	Add(words ...string)
	// Remove 批量移除关键字；不存在的词静默跳过。
	Remove(words ...string)
	// Match 返回命中的关键字；默认匹配第一个，未命中返回 false 与空串。
	Match(text string, opts ...MatchOption) (found bool, word string)
}

// Filter 高层敏感词过滤器：在 Trie 之上提供去噪匹配能力。
type Filter interface {
	// AddWord 批量添加敏感词。
	AddWord(words ...string)
	// RemoveWord 批量移除敏感词；不存在的词静默跳过。
	RemoveWord(words ...string)
	// Match 返回命中的敏感词；默认匹配第一个，可用 WithRemoveNoise 控制匹配前是否先去噪。
	Match(text string, opts ...MatchOption) (found bool, word string)
}

var (
	_ Trie   = (*trie)(nil)
	_ Filter = (*filter)(nil)
)
