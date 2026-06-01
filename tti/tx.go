package tti

import (
	"context"
	"encoding/json"

	"github.com/advancevillage/3rd/dbx"
	"github.com/advancevillage/3rd/logx"
	"github.com/advancevillage/3rd/mathx"
	"github.com/advancevillage/3rd/netx"
	"github.com/advancevillage/3rd/x"
	"github.com/advancevillage/xmagic/pkg/pb"
)

var _ AiGenerator = (*imagen)(nil)

type imagen struct {
	s3        dbx.S3
	opts      generateOption
	path      string
	ider      mathx.IDGenerator
	logger    logx.ILogger
	endpoint  string
	transport netx.HttpClient
}

func NewStabilityClient(ctx context.Context, logger logx.ILogger, s3 dbx.S3, opt ...GenerateOption) (AiGenerator, error) {
	return newStabilityClient(ctx, logger, s3, opt...)
}

func newStabilityClient(ctx context.Context, logger logx.ILogger, s3 dbx.S3, opt ...GenerateOption) (*imagen, error) {
	// 1. 解析参数
	opts := defaultGenerateOptions
	for _, o := range opt {
		o.apply(&opts)
	}
	// 2. 发号器
	ider, err := mathx.NewSnowFlake(678)
	if err != nil {
		logger.Errorw(ctx, "imagen snowflake failed", "err", err)
		return nil, err
	}
	// 3. 创建请求对象
	transport, err := netx.NewHttpClient(ctx, logger,
		netx.WithClientHeader(
			x.WithKV("authorization", "Bearer "+opts.token),
		),
		netx.WithClientTimeout(300),
	)
	if err != nil {
		logger.Errorw(ctx, "new imagen client transport", "err", err)
		return nil, err
	}
	return &imagen{
		s3:        s3,
		opts:      opts,
		ider:      ider,
		path:      "/v1/api/image/submit",
		logger:    logger,
		endpoint:  "https://tokenhub.tencentmaas.com",
		transport: transport,
	}, nil
}

func (c *imagen) Generate(ctx context.Context, prompt string, opts ...x.Option) (pb.Descriptor, error) {
	d, err := newDescriptor(ctx, c.logger, c.s3, c.ider, WithGeneratePrefix("hy"))
	if err != nil {
		c.logger.Errorw(ctx, "imagen descriptor failed", "err", err)
		return nil, err
	}
	go c.asyncGenerate(ctx, d, prompt, opts...)
	return d, nil
}

func (c *imagen) asyncGenerate(ctx context.Context, d *descriptor, prompt string, opts ...x.Option) {
	r, err := c.generate(ctx, prompt, opts...)
	if err != nil {
		d.errCh <- err
	} else {
		d.reply <- r
	}
	d.clear()
}

func (c *imagen) generate(ctx context.Context, prompt string, opts ...x.Option) (netx.HttpResponse, error) {
	hdr := x.NewBuilder(
		x.WithKV("Content-Type", "application/json"),
	)
	body, err := json.Marshal(x.NewBuilder(
		x.WithKV("model", "hy-image-v3.0"),
		x.WithKV("prompt", prompt),
	).Build())
	if err != nil {
		return nil, err
	}
	res, err := c.transport.Post(ctx, c.endpoint+c.path, hdr, body)
	if err != nil {
		return nil, err
	}
	c.logger.Infow(ctx, "imagen image generate success", "res", res)
	return res, nil
}
