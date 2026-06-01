package tti

import (
	"context"
	"time"

	"github.com/advancevillage/3rd/netx"
)

type GenerateOption interface {
	apply(*generateOption)
}

type GenerateParser func(ctx context.Context, reply netx.HttpResponse) (netx.HttpResponse, error)

func WitGenerateSecret(sk string) GenerateOption {
	return newFuncGenerateOption(func(o *generateOption) {
		o.token = sk
	})
}

func WithGeneratePrefix(prefix string) GenerateOption {
	return newFuncGenerateOption(func(o *generateOption) {
		o.prefix = prefix
	})
}

func WithGenerateParser(parser GenerateParser) GenerateOption {
	return newFuncGenerateOption(func(o *generateOption) {
		o.parser = parser
	})
}

func WithGenerateProxy(proxy string) GenerateOption {
	return newFuncGenerateOption(func(o *generateOption) {
		o.proxy = proxy
	})
}

func emptyGenerateParser(ctx context.Context, reply netx.HttpResponse) (netx.HttpResponse, error) {
	return reply, nil
}

type generateOption struct {
	ext     string         // 扩展名
	token   string         // 密钥
	proxy   string         // 代理
	prefix  string         // 前缀
	parser  GenerateParser // 解析器
	timeout time.Duration  // 超时时间
}

var defaultGenerateOptions = generateOption{
	ext:     ".png",
	prefix:  "tti",
	parser:  emptyGenerateParser,
	timeout: time.Minute * 5,
}

type funcGenerateOption struct {
	f func(*generateOption)
}

func (fdo *funcGenerateOption) apply(do *generateOption) {
	fdo.f(do)
}

func newFuncGenerateOption(f func(*generateOption)) *funcGenerateOption {
	return &funcGenerateOption{
		f: f,
	}
}
