package mcp

import "io"

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

type ServerOption = Option[serverOption]

type serverOption struct {
	name    string
	version string
	reader  io.Reader
	writer  io.Writer
}

var defaultServerOptions = serverOption{
	name:    "mcp-server",
	version: "1.0.0",
}

func WithServerName(name string) ServerOption {
	return newFuncOption(func(o *serverOption) {
		o.name = name
	})
}

func WithServerVersion(version string) ServerOption {
	return newFuncOption(func(o *serverOption) {
		o.version = version
	})
}

func WithReader(r io.Reader) ServerOption {
	return newFuncOption(func(o *serverOption) {
		o.reader = r
	})
}

func WithWriter(w io.Writer) ServerOption {
	return newFuncOption(func(o *serverOption) {
		o.writer = w
	})
}
