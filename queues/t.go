//author: richard
package queues


func NewTQueue(count int) *TQueue {
	t := &TQueue{
		count: count,
		size: 0,
		head: 0,
		tail: 0,
		q : make([]*Element, count),
	}
	return t
}

func (t *TQueue) Empty() bool {
	return t.size == 0
}

func (t *TQueue) Full() bool {
	return t.size == t.count
}

func (t *TQueue) Len() int {
	return t.size
}

func (t *TQueue) Enqueue(e *Element) error {
	if t.Full() {
		return QueueFullError
	}
	t.q[t.tail % t.count] = e
	t.size++
	t.tail = (t.tail + 1) % t.count
	return nil
}

func (t *TQueue) Dequeue() (*Element, error) {
	if t.Empty() {
		return nil, QueueEmptyError
	}
	e := t.q[t.head % t.count]
	t.size--
	t.head = (t.head + 1) % t.count
	return e, nil
}