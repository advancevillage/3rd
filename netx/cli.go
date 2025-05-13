package netx

import (
	"context"
	"net/http"

	"github.com/advancevillage/3rd/logx"
	"github.com/advancevillage/3rd/x"
	"google.golang.org/grpc"
)

type GrpcClient interface {
	Conn(ctx context.Context) grpc.ClientConnInterface
	Close(ctx context.Context)
}

func NewGrpcClient(ctx context.Context, logger logx.ILogger, opt ...ClientOption) (GrpcClient, error) {
	return newGrpcCli(ctx, logger, opt...)
}

func NewHttpClient(ctx context.Context, logger logx.ILogger, opt ...ClientOption) (HttpClient, error) {
	return newHttpClient(ctx, logger, opt...)
}

type HttpResponse interface {
	Body() []byte
	Header() http.Header
	StatusCode() int
}

type ClientOption interface {
	apply(*clientOptions)
}

func WithClientAddr(host string, port int) ClientOption {
	return newFuncClientOption(func(o *clientOptions) {
		o.host = host
		o.port = port
	})
}

func WithClientTimeout(timeout int) ClientOption {
	return newFuncClientOption(func(o *clientOptions) {
		o.timeout = timeout
	})
}

func WithClientHeader(hdr ...x.Option) ClientOption {
	return newFuncClientOption(func(o *clientOptions) {
		b := x.NewBuilder(hdr...)
		o.hdr = b.Build()
	})
}

func WithClientCredential(crt, domain string) ClientOption {
	return newFuncClientOption(func(o *clientOptions) {
		o.crt = crt
		o.domain = domain
	})
}

type clientOptions struct {
	hdr     map[string]interface{} // 请求头
	host    string                 // 服务地址
	port    int                    // 端口
	crt     string                 // 证书 文件
	domain  string                 // 域名
	timeout int
}

var defaultClientOptions = clientOptions{
	hdr:     map[string]interface{}{},
	host:    "127.0.0.1",
	port:    13147,
	crt:     "cert.pem",
	domain:  "api.softpart.cn",
	timeout: 3,
}

type funcClientOption struct {
	f func(*clientOptions)
}

func (fdo *funcClientOption) apply(do *clientOptions) {
	fdo.f(do)
}

func newFuncClientOption(f func(*clientOptions)) *funcClientOption {
	return &funcClientOption{
		f: f,
	}
}
