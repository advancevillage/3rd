package x

type Options[T any] interface {
	Apply(*T)
}

type funcOptions[T any] struct {
	f func(*T)
}

func (o funcOptions[T]) Apply(do *T) {
	o.f(do)
}

func NewFuncOptions[T any](f func(*T)) *funcOptions[T] {
	return &funcOptions[T]{f: f}
}
