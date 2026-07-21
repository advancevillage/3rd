package trie

// node Trie 树上的一个节点。
type node struct {
	isRootNode bool
	isPathEnd  bool
	character  rune
	children   map[rune]*node
	failure    *node
	parent     *node
	values     map[string]struct{}
	depth      int
}

// trie 短语组成的 Trie 树。
type trie struct {
	root *node
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
		values:    make(map[string]struct{}),
	}
}

// newRootNode 新建根节点。
func newRootNode(character rune) *node {
	root := &node{
		isRootNode: true,
		character:  character,
		children:   make(map[rune]*node),
		values:     make(map[string]struct{}),
	}
	root.failure = root
	return root
}

// BuildFailureLinks 构建 Aho-Corasick 失败指针。
func (t *trie) BuildFailureLinks() {
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
}

// AddWithValues 添加关键字并关联一组自定义值。
func (t *trie) AddWithValues(word string, values []string) {
	current := t.root
	runes := []rune(word)
	for position := 0; position < len(runes); position++ {
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
			// 词末尾节点添加关联值
			for _, value := range values {
				current.values[value] = struct{}{}
			}
		}
	}
}

func (t *trie) add(word string) {
	current := t.root
	runes := []rune(word)
	for position := 0; position < len(runes); position++ {
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

// Replace 将命中的关键字替换为 character。
func (t *trie) Replace(text string, character rune) string {
	n := t.root
	runes := []rune(text)
	a := new(ac)
	for position := 0; position < len(runes); position++ {
		next := a.next(n, runes[position])
		if next == nil {
			next = a.fail(n, runes[position])
		}
		n = next
		a.replace(n, runes, position, character)
	}
	return string(runes)
}

// Filter 直接删除文本中命中的关键字。
func (t *trie) Filter(text string) string {
	parent := t.root
	var current *node
	left := 0
	var found bool
	runes := []rune(text)
	length := len(runes)
	resultRunes := make([]rune, 0, length)

	for position := 0; position < length; position++ {
		current, found = parent.children[runes[position]]
		if !found {
			resultRunes = append(resultRunes, runes[left])
			parent = t.root
			position = left
			left++
			continue
		}
		if current.isPathEnd {
			left = position + 1
		}
		parent = current
	}

	resultRunes = append(resultRunes, runes[left:]...)
	return string(resultRunes)
}

// Validate 校验文本是否合法；不合法返回 false 及命中的第一个关键字。
func (t *trie) Validate(text string) (bool, string) {
	const empty = ""
	n := t.root
	runes := []rune(text)
	a := new(ac)
	for position := 0; position < len(runes); position++ {
		next := a.next(n, runes[position])
		if next == nil {
			next = a.fail(n, runes[position])
		}
		n = next
		if first := a.firstOutput(n, runes, position); len(first) > 0 {
			return false, first
		}
	}
	return true, empty
}

// FindIn 判断文本中是否含有词库中的词。
func (t *trie) FindIn(text string) (bool, string) {
	validated, first := t.Validate(text)
	return !validated, first
}

// MatchFirst 返回第一个命中的关键字，未命中返回空字符串。
func (t *trie) MatchFirst(text string) (bool, string) {
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

// BatchMatchFirst 批量匹配，返回第一个命中的关键字。
func (t *trie) BatchMatchFirst(texts []string) (bool, string) {
	for _, txt := range texts {
		if matched, word := t.MatchFirst(txt); matched {
			return true, word
		}
	}
	return false, ""
}

// MatchFirstAndValues 返回第一个命中的关键字及其关联值。
func (t *trie) MatchFirstAndValues(text string) (bool, string, []string) {
	n := t.root
	runes := []rune(text)
	for i, r := range runes {
		for n != t.root && n.children[r] == nil {
			n = n.failure
		}
		if next, ok := n.children[r]; ok {
			n = next
		} else {
			n = t.root
		}
		if n.isPathEnd {
			start := i + 1 - n.depth
			values := make([]string, 0, len(n.values))
			for v := range n.values {
				values = append(values, v)
			}
			return true, string(runes[start : i+1]), values
		}
	}
	return false, "", nil
}

// BatchMatchFirstAndValues 批量匹配，返回第一个命中的关键字及其关联值。
func (t *trie) BatchMatchFirstAndValues(texts []string) (bool, string, []string) {
	for _, txt := range texts {
		if matched, word, values := t.MatchFirstAndValues(txt); matched {
			return true, word, values
		}
	}
	return false, "", nil
}

// FindAll 找到所有包含在词库中的词。
func (t *trie) FindAll(text string) []string {
	n := t.root
	runes := []rune(text)
	a := new(ac)
	for position := 0; position < len(runes); position++ {
		next := a.next(n, runes[position])
		if next == nil {
			next = a.fail(n, runes[position])
		}
		n = next
		a.output(n, runes, position)
	}
	return a.results
}
