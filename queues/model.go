//author: richard
package queues

import "errors"

var (
	QueueFullError  = errors.New("queue is full")
	QueueEmptyError = errors.New("queue is empty")
)

type Queue interface {
	Enqueue(e *Element) error
	Dequeue() (*Element, error)
	Empty() bool
	Full()  bool
	Len()   int
}

type Element struct {
	Key  string
	Body []byte
}

type TQueue struct {
	count int
	size  int
	head  int
	tail  int
	q	  []*Element
}