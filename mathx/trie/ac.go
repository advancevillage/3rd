package trie

// ac 是 Aho-Corasick 自动机的匹配辅助器。
type ac struct {
	results []string
}

func (a *ac) fail(n *node, c rune) *node {
	for {
		next := a.next(n.failure, c)
		if next == nil {
			if n.isRootNode {
				return n
			}
			n = n.failure
			continue
		}
		return next
	}
}

func (a *ac) next(n *node, c rune) *node {
	if next, ok := n.children[c]; ok {
		return next
	}
	return nil
}

func (a *ac) output(n *node, runes []rune, position int) {
	if n.isRootNode {
		return
	}
	if n.isPathEnd {
		a.results = append(a.results, string(runes[position+1-n.depth:position+1]))
	}
	a.output(n.failure, runes, position)
}

func (a *ac) firstOutput(n *node, runes []rune, position int) string {
	if n.isRootNode {
		return ""
	}
	if n.isPathEnd {
		return string(runes[position+1-n.depth : position+1])
	}
	return a.firstOutput(n.failure, runes, position)
}

func (a *ac) replace(n *node, runes []rune, position int, replace rune) {
	if n.isRootNode {
		return
	}
	if n.isPathEnd {
		for i := position + 1 - n.depth; i < position+1; i++ {
			runes[i] = replace
		}
	}
	a.replace(n.failure, runes, position, replace)
}
