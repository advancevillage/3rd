package tti

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/advancevillage/3rd/dbx"
	"github.com/advancevillage/3rd/logx"
	"github.com/advancevillage/3rd/mathx"
	"github.com/advancevillage/3rd/netx"
	"github.com/advancevillage/3rd/x"
)

var _ AiGenerator = (*seedream)(nil)

// ByteDance Seedream Image Model
type seedream struct {
	s3         dbx.S3
	opts       generateOption
	submitPath string
	ider       mathx.IDGenerator
	logger     logx.ILogger
	endpoint   string
	transport  netx.HttpClient
	downloader netx.HttpClient
}

func NewBdImageClient(ctx context.Context, logger logx.ILogger, s3 dbx.S3, opt ...GenerateOption) (AiGenerator, error) {
	return newBdImageClient(ctx, logger, s3, opt...)
}

func newBdImageClient(ctx context.Context, logger logx.ILogger, s3 dbx.S3, opt ...GenerateOption) (*seedream, error) {
	// 1. 解析参数
	opts := defaultGenerateOptions
	for _, o := range opt {
		o.apply(&opts)
	}
	// 2. 发号器
	ider, err := mathx.NewSnowFlake(678)
	if err != nil {
		logger.Errorw(ctx, "seedream snowflake failed", "err", err)
		return nil, err
	}
	// 3. 创建请求对象 (鉴权)
	transport, err := netx.NewHttpClient(ctx, logger,
		netx.WithClientHeader(
			x.WithKV("authorization", "Bearer "+opts.token),
		),
		netx.WithClientTimeout(300),
	)
	if err != nil {
		logger.Errorw(ctx, "new seedream client transport", "err", err)
		return nil, err
	}
	// 4. 下载图片的客户端
	downloader, err := netx.NewHttpClient(ctx, logger,
		netx.WithClientTimeout(300),
	)
	if err != nil {
		logger.Errorw(ctx, "new seedream downloader transport", "err", err)
		return nil, err
	}
	return &seedream{
		s3:         s3,
		opts:       opts,
		ider:       ider,
		submitPath: "/api/v3/images/generations",
		logger:     logger,
		endpoint:   "https://ark.cn-beijing.volces.com",
		transport:  transport,
		downloader: downloader,
	}, nil
}

func (c *seedream) Generate(ctx context.Context, prompt string, opts ...x.Option) (Descriptor, error) {
	d, err := newDescriptor(ctx, c.logger, c.s3, c.ider, WithGeneratePrefix(c.opts.prefix))
	if err != nil {
		c.logger.Errorw(ctx, "seedream descriptor failed", "err", err)
		return nil, err
	}
	go c.asyncGenerate(ctx, d, prompt, opts...)
	return d, nil
}

func (c *seedream) asyncGenerate(ctx context.Context, d *descriptor, prompt string, opts ...x.Option) {
	r, err := c.generate(ctx, prompt, opts...)
	if err != nil {
		d.errCh <- err
	} else {
		d.reply <- r
	}
	d.clear()
}

func (c *seedream) generate(ctx context.Context, prompt string, opts ...x.Option) (netx.HttpResponse, error) {
	hdr := x.NewBuilder(
		x.WithKV("Content-Type", "application/json"),
	)
	body, err := json.Marshal(x.NewBuilder(
		x.WithKV("model", c.opts.model),
		x.WithKV("prompt", prompt+";aspect_ratio:"+c.opts.aspectRatio),
		x.WithKV("output_format", strings.Trim(c.opts.ext, ".")),
		x.WithKV("size", "2K"),
		x.WithKV("watermark", false),
		x.WithKV("sequential_image_generation", "disabled"),
		x.WithKV("response_format", "url"),
	).Build())
	if err != nil {
		return nil, err
	}
	res, err := c.transport.Post(ctx, c.endpoint+c.submitPath, hdr, body)
	if err != nil {
		return nil, err
	}
	c.logger.Infow(ctx, "seedream image generate success", "res", string(res.Body()))

	var sub bdSubmitResp
	if err = json.Unmarshal(res.Body(), &sub); err != nil {
		return nil, err
	}
	if len(sub.Data) == 0 || sub.Data[0] == nil || sub.Data[0].Url == "" {
		return nil, errors.New("seedream image generate failed, empty url")
	}
	return c.downloader.Get(ctx, sub.Data[0].Url, x.NewBuilder(), x.NewBuilder())
}
