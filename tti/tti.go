package tti

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/advancevillage/3rd/dbx"
	"github.com/advancevillage/3rd/logx"
	"github.com/advancevillage/3rd/mathx"
	"github.com/advancevillage/3rd/netx"
	"github.com/advancevillage/3rd/x"
)

type Descriptor interface {
	Id(ctx context.Context) string
	Name(ctx context.Context) (string, error)
	Progress(ctx context.Context) (int32, error)
	Resource(ctx context.Context) (string, error)
}

type AiGenerator interface {
	Generate(ctx context.Context, promptOrUrl string, opts ...x.Option) (Descriptor, error)
}

var _ Descriptor = (*descriptor)(nil)

type descriptor struct {
	s3       dbx.S3
	sid      string
	err      error
	opts     generateOption
	name     string
	time     time.Time
	logger   logx.ILogger
	progress int32

	errCh chan error
	reply chan netx.HttpResponse
}

func newDescriptor(ctx context.Context, logger logx.ILogger, s3 dbx.S3, ider mathx.IDGenerator, opt ...GenerateOption) (*descriptor, error) {
	// 1. 解析参数
	opts := defaultGenerateOptions
	for _, o := range opt {
		o.apply(&opts)
	}
	// 2. 构建
	d := &descriptor{
		s3:       s3,
		sid:      mathx.UUID(),
		err:      nil,
		opts:     opts,
		name:     fmt.Sprintf("%s%d%s", strings.ToUpper(opts.prefix), ider.Generate(), opts.ext),
		time:     time.Now(),
		logger:   logger,
		progress: 0,

		errCh: make(chan error, 1),
		reply: make(chan netx.HttpResponse, 1),
	}
	// 3. 监听
	go d.loop(ctx)
	return d, nil
}

func (c *descriptor) loop(ctx context.Context) {
	sctx, cancel := context.WithTimeout(ctx, c.opts.timeout)
	defer cancel()

	t := time.NewTicker(2 * time.Second)
	defer t.Stop()

	for {
		c.logger.Infow(ctx, "watch generate image", "name", c.name, "sid", c.sid, "progress", c.progress)
		select {
		case <-sctx.Done():
			c.err = sctx.Err()
			goto exitLoop

		case <-t.C:
			pg := mathx.SmoothShiftedSigmoid(float64(time.Since(c.time)/time.Second), 0.08, 60)
			c.progress = int32(pg * 100)

		case c.err = <-c.errCh:
			c.logger.Infow(ctx, "watch generate image failed", "name", c.name, "sid", c.sid, "progress", c.progress, "err", c.err)
			goto exitLoop

		case r := <-c.reply:
			r, err := c.opts.parser(sctx, r)
			if err != nil {
				c.err = err
				goto exitLoop
			}
			err = x.Upload(ctx, c.logger, c.s3, c.name, r.Body())
			if err != nil {
				c.err = err
				goto exitLoop
			}
			c.progress = 100
			goto exitLoop
		}
	}
exitLoop:
	c.logger.Infow(ctx, "watch generate image exit", "name", c.name, "sid", c.sid, "progress", c.progress, "err", c.err)
}

func (c *descriptor) clear() {
	close(c.errCh)
	close(c.reply)
}

func (c *descriptor) Id(ctx context.Context) string {
	return c.sid
}

func (c *descriptor) Resource(ctx context.Context) (string, error) {
	var (
		uri string
		err error
	)
	if c.err != nil {
		return uri, c.err
	}
	uri, err = c.s3.Url(ctx, c.name)
	if err != nil {
		return uri, err
	}
	return uri, nil
}

func (c *descriptor) Progress(ctx context.Context) (int32, error) {
	if c.progress >= 100 || c.err != nil {
		return c.progress, c.err
	}
	pg := mathx.SmoothShiftedSigmoid(float64(time.Since(c.time)/time.Second), 0.08, 60)
	return int32(pg * 100), nil
}

func (c *descriptor) Name(ctx context.Context) (string, error) {
	return c.name, c.err
}
