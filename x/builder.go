package x

import (
	"fmt"
	"strings"
	"sync"
)

func WithKV(key string, value interface{}) Option {
	return func(b Builder) {
		b.write(strings.TrimSpace(key), value)
	}
}

type Builder interface {
	Value(key string) (any, bool)
	Build() map[string]interface{}
	write(key string, value interface{})
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

func (o *builder) write(key string, value interface{}) {
	o.m.Store(key, value)
}

func (o *builder) Build() map[string]interface{} {
	m := make(map[string]interface{})
	o.m.Range(func(key, value interface{}) bool {
		m[fmt.Sprint(key)] = value
		o.m.Delete(key)
		return true
	})
	return m
}

func (o *builder) Value(key string) (any, bool) {
	return o.m.Load(key)
}
