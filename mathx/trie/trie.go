package trie

// node Trie 树上的一个节点。
type node struct {
	isRootNode bool
	isPathEnd  bool
	character  rune
	children   map[rune]*node
	failure    *node
	parent     *node
	depth      int
}

// trie 短语组成的 Trie 树。
type trie struct {
	root  *node
	dirty bool // 词库变更后需（重新）构建失败指针
}

// NewTrie 新建一棵空 Trie。
func NewTrie() (Trie, error) {
	return newTrie(), nil
}

func newTrie() *trie {
	return &trie{root: newRootNode(0)}
}

// newNode 新建子节点。
func newNode(character rune) *node {
	return &node{
		character: character,
		children:  make(map[rune]*node),
	}
}

// newRootNode 新建根节点。
func newRootNode(character rune) *node {
	root := &node{
		isRootNode: true,
		character:  character,
		children:   make(map[rune]*node),
	}
	root.failure = root
	return root
}

// ensureLinks 在匹配前懒构建失败指针；词库变更后首次匹配才会真正重建。
func (t *trie) ensureLinks() {
	if t.dirty {
		t.buildFailureLinks()
		t.dirty = false
	}
}

// buildFailureLinks 构建 Aho-Corasick 失败指针。
func (t *trie) buildFailureLinks() {
	for n := range t.bfs() {
		pointer := n.parent
		var link *node
		for link == nil {
			if pointer.isRootNode {
				link = pointer
				break
			}
			link = pointer.failure.children[n.character]
			pointer = pointer.failure
		}
		n.failure = link
	}
}

// bfs 广度优先遍历树上的每个节点（不含根节点）。
func (t *trie) bfs() <-chan *node {
	ch := make(chan *node)
	go func() {
		queue := new(linkList)
		for _, child := range t.root.children {
			queue.push(child)
		}
		for !queue.empty() {
			n := queue.pop().(*node)
			ch <- n
			for _, child := range n.children {
				queue.push(child)
			}
		}
		close(ch)
	}()
	return ch
}

// Add 批量添加关键字。
func (t *trie) Add(words ...string) {
	for _, word := range words {
		t.add(word)
	}
	t.dirty = true
}

// Remove 批量移除关键字；不存在的词静默跳过。
func (t *trie) Remove(words ...string) {
	for _, word := range words {
		t.remove(word)
	}
}

func (t *trie) remove(word string) {
	runes := []rune(word)
	if len(runes) == 0 {
		return
	}
	// 沿路径下行定位词尾节点；任一字符缺失说明该词不存在。
	current := t.root
	for _, r := range runes {
		next, ok := current.children[r]
		if !ok {
			return
		}
		current = next
	}
	if !current.isPathEnd {
		return
	}
	// 清除词尾标记。
	current.isPathEnd = false
	// 自底向上剪掉「无子节点且非其它词词尾」的节点，避免误删公共前缀。
	for !current.isRootNode && !current.isPathEnd && len(current.children) == 0 {
		parent := current.parent
		delete(parent.children, current.character)
		current = parent
	}
	// 结构变化使失败指针整体失效，下次匹配重建。
	t.dirty = true
}

func (t *trie) add(word string) {
	current := t.root
	runes := []rune(word)
	for position := range len(runes) {
		r := runes[position]
		if next, ok := current.children[r]; ok {
			current = next
		} else {
			child := newNode(r)
			child.depth = current.depth + 1
			child.parent = current
			current.children[r] = child
			current = child
		}
		if position == len(runes)-1 {
			current.isPathEnd = true
		}
	}
}

// Match 返回命中的关键字，默认匹配第一个；未命中返回 false 与空串。
func (t *trie) Match(text string, _ ...MatchOption) (bool, string) {
	t.ensureLinks()
	n := t.root
	runes := []rune(text)
	for i, r := range runes {
		// 按 fail 指针回溯，直到找到能匹配的边或回到 root
		for n != t.root && n.children[r] == nil {
			n = n.failure
		}
		if next, ok := n.children[r]; ok {
			n = next
		} else {
			n = t.root
		}
		// 如果该节点是某个关键字的结尾，直接提前返回
		if n.isPathEnd {
			start := i + 1 - n.depth
			return true, string(runes[start : i+1])
		}
	}
	return false, ""
}
