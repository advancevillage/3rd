package tti

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/advancevillage/3rd/dbx"
	"github.com/advancevillage/3rd/logx"
	"github.com/advancevillage/3rd/mathx"
	"github.com/advancevillage/3rd/netx"
	"github.com/advancevillage/3rd/x"
)

var _ AiGenerator = (*imagen)(nil)

const (
	txStatusCompleted = "completed"
	txStatusFailed    = "failed"
	txPollInterval    = 2 * time.Second
)

type imagen struct {
	s3         dbx.S3
	opts       generateOption
	submitPath string
	queryPath  string
	ider       mathx.IDGenerator
	logger     logx.ILogger
	endpoint   string
	transport  netx.HttpClient
	downloader netx.HttpClient
}

func NewTxImageClient(ctx context.Context, logger logx.ILogger, s3 dbx.S3, opt ...GenerateOption) (AiGenerator, error) {
	return newTxImageClient(ctx, logger, s3, opt...)
}

func newTxImageClient(ctx context.Context, logger logx.ILogger, s3 dbx.S3, opt ...GenerateOption) (*imagen, error) {
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
	// 3. 创建请求对象 (鉴权)
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
	// 4. 下载图片的客户端 (不携带鉴权头, 避免向 COS 泄漏 SK)
	downloader, err := netx.NewHttpClient(ctx, logger,
		netx.WithClientTimeout(300),
	)
	if err != nil {
		logger.Errorw(ctx, "new imagen downloader transport", "err", err)
		return nil, err
	}
	return &imagen{
		s3:         s3,
		opts:       opts,
		ider:       ider,
		submitPath: "/v1/api/image/submit",
		queryPath:  "/v1/api/image/query",
		logger:     logger,
		endpoint:   "https://tokenhub.tencentmaas.com",
		transport:  transport,
		downloader: downloader,
	}, nil
}

func (c *imagen) Generate(ctx context.Context, prompt string, opts ...x.Option) (Descriptor, error) {
	d, err := newDescriptor(ctx, c.logger, c.s3, c.ider, WithGeneratePrefix(c.opts.prefix))
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
		x.WithKV("model", c.opts.model),
		x.WithKV("prompt", prompt),
		x.WithKV("aspect_ratio", c.opts.aspectRatio),
	).Build())
	if err != nil {
		return nil, err
	}
	res, err := c.transport.Post(ctx, c.endpoint+c.submitPath, hdr, body)
	if err != nil {
		return nil, err
	}
	c.logger.Infow(ctx, "imagen image generate success", "res", string(res.Body()))

	var sub txSubmitResp
	if err = json.Unmarshal(res.Body(), &sub); err != nil {
		return nil, err
	}
	if sub.Id == "" {
		return nil, fmt.Errorf("imagen submit empty id, resp=%s", string(res.Body()))
	}

	url, err := c.poll(ctx, sub.Id)
	if err != nil {
		return nil, err
	}
	return c.downloader.Get(ctx, url, x.NewBuilder(), x.NewBuilder())
}

func (c *imagen) poll(ctx context.Context, id string) (string, error) {
	ticker := time.NewTicker(txPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-ticker.C:
			q, err := c.query(ctx, id)
			if err != nil {
				return "", err
			}
			c.logger.Infow(ctx, "imagen image query", "id", id, "status", q.Status)
			switch q.Status {
			case txStatusCompleted:
				if len(q.Data) == 0 {
					return "", fmt.Errorf("imagen query completed but no data, id=%s", id)
				}
				return q.Data[0].Url, nil
			case txStatusFailed:
				return "", fmt.Errorf("imagen query failed, id=%s, request_id=%s", id, q.RequestId)
			}
		}
	}
}

func (c *imagen) query(ctx context.Context, id string) (*txQueryResp, error) {
	hdr := x.NewBuilder(
		x.WithKV("Content-Type", "application/json"),
	)
	body, err := json.Marshal(x.NewBuilder(
		x.WithKV("model", c.opts.model),
		x.WithKV("id", id),
	).Build())
	if err != nil {
		return nil, err
	}
	res, err := c.transport.Post(ctx, c.endpoint+c.queryPath, hdr, body)
	if err != nil {
		return nil, err
	}
	var q txQueryResp
	if err = json.Unmarshal(res.Body(), &q); err != nil {
		return nil, err
	}
	return &q, nil
}
