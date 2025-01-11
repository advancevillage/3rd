package netx

import (
	"context"

	"github.com/advancevillage/3rd/logx"
	"google.golang.org/grpc"
)

type GrpcClient interface {
	Conn(ctx context.Context) grpc.ClientConnInterface
	Close(ctx context.Context)
}

func NewGrpcClient(ctx context.Context, logger logx.ILogger, opt ...ClientOption) (GrpcClient, error) {
	return newGrpcCli(ctx, logger, opt...)
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

func WithClientCredential(crt, domain string) ClientOption {
	return newFuncClientOption(func(o *clientOptions) {
		o.crt = crt
		o.domain = domain
	})
}

type clientOptions struct {
	host   string // 服务地址
	port   int    // 端口
	crt    string // 证书 文件
	domain string // 域名
}

var defaultClientOptions = clientOptions{
	host:   "127.0.0.1",
	port:   13147,
	crt:    "cert.pem",
	domain: "api.sunhe.org",
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
