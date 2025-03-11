package x

import (
	"fmt"
	"strings"
	"sync"
)

func WithKV(key string, value any) Option {
	return func(b Builder) {
		b.write(strings.TrimSpace(key), value)
	}
}

type Builder interface {
	Value(key string) (any, bool)
	Build() map[string]any
	write(key string, value any)
}

type Option func(Builder)

var _ Builder = (*builder)(nil)

type builder struct {
	m sync.Map
}

func NewBuilder(opts ...Option) Builder {
	b := &builder{}
	for _, opt := range opts {
		opt(b)
	}
	return b
}

func (o *builder) write(key string, value any) {
	o.m.Store(key, value)
}

func (o *builder) Build() map[string]any {
	m := make(map[string]any)
	o.m.Range(func(key, value any) bool {
		m[fmt.Sprint(key)] = value
		o.m.Delete(key)
		return true
	})
	return m
}

func (o *builder) Value(key string) (any, bool) {
	return o.m.Load(key)
}
