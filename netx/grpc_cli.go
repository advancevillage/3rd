package netx

import (
	"fmt"

	"github.com/advancevillage/3rd/logx"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var _ GrpcClient = (*grpcCli)(nil)

type grpcCli struct {
	opts   clientOptions
	logger logx.ILogger
	conn   *grpc.ClientConn
}

func newGrpcCli(ctx context.Context, logger logx.ILogger, opt ...ClientOption) (*grpcCli, error) {
	// 0. 设置配置
	opts := defaultClientOptions
	for _, o := range opt {
		o.apply(&opts)
	}

	c := &grpcCli{logger: logger, opts: opts}

	// 1. 安全凭证
	creds, err := credentials.NewClientTLSFromFile(c.opts.crt, c.opts.domain)
	if err != nil {
		c.logger.Errorw(ctx, "read cert file failed", "err", err, "crt", c.opts.crt, "domain", c.opts.domain)
		return nil, err
	}

	// 2. 连接拨号
	c.conn, err = grpc.NewClient(fmt.Sprintf("%s:%d", c.opts.host, c.opts.port), grpc.WithTransportCredentials(creds))
	if err != nil {
		c.logger.Errorw(ctx, "grpc dial failed", "err", err, "host", c.opts.host, "port", c.opts.port)
		return nil, err
	}

	return c, nil
}

func (c *grpcCli) Conn(ctx context.Context) grpc.ClientConnInterface {
	return c.conn
}

func (c *grpcCli) Close(ctx context.Context) {
	err := c.conn.Close()
	if err != nil {
		c.logger.Errorw(ctx, "grpc close failed", "err", err)
	}
}
