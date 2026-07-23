package trie

// linkList 是 BFS 遍历使用的先进先出队列。
type linkList struct {
	head  *listNode
	tail  *listNode
	count int64
}

type listNode struct {
	value any
	next  *listNode
}

// push 入队。
func (l *linkList) push(v any) {
	n := &listNode{value: v}
	if l.head == nil {
		l.head = n
	} else {
		l.tail.next = n
	}
	l.tail = n
	l.count++
}

// pop 出队并返回队首元素的值；队列为空时返回 nil。
func (l *linkList) pop() any {
	if l.empty() {
		return nil
	}
	n := l.head
	l.head = n.next
	l.count--
	return n.value
}

// empty 队列是否为空。
func (l *linkList) empty() bool {
	return l.count == 0
}
