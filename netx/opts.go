package netx

type Option[T any] interface {
	apply(*T)
}

type funcOption[T any] struct {
	f func(*T)
}

func (o funcOption[T]) apply(do *T) {
	o.f(do)
}

func newFuncOption[T any](f func(*T)) *funcOption[T] {
	return &funcOption[T]{f: f}
}
