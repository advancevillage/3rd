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
	"golang.org/x/sync/errgroup"
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

// 进度曲线参数：sigmoid 中点 x0 秒、陡度 alpha
//
//	t=0  → ~5%   t=10 → ~18%   t=20 → 50%   t=30 → ~82%   t=40 → ~95%
const (
	progressAlpha = 0.15
	progressX0    = 20
)

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
		name:     fmt.Sprintf("%s%d%s", strings.ToLower(opts.prefix), ider.Generate(), opts.ext),
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

	t := time.NewTicker(time.Second * 2)
	defer t.Stop()

	for {
		c.logger.Infow(ctx, "watch generate image", "name", c.name, "sid", c.sid, "progress", c.progress)
		select {
		case <-sctx.Done():
			c.err = sctx.Err()
			goto exitLoop

		case <-t.C:
			pg := mathx.SmoothShiftedSigmoid(float64(time.Since(c.time)/time.Second), progressAlpha, progressX0)
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
			err = c.upload(ctx, r.Body())
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
	pg := mathx.SmoothShiftedSigmoid(float64(time.Since(c.time)/time.Second), progressAlpha, progressX0)
	return int32(pg * 100), nil
}

func (c *descriptor) Name(ctx context.Context) (string, error) {
	return c.name, c.err
}

func (c *descriptor) upload(ctx context.Context, body []byte) error {
	var (
		n     = len(body)
		part  = 0
		chunk = 1 << 20
		total = n / chunk
	)
	if n%chunk > 0 {
		total += 1
	}
	c.logger.Infow(ctx, "upload info", "name", c.name, "size", n, "total", total, "chunk", chunk)

	u, err := c.s3.MultiUpload(ctx, c.name, total, dbx.WithContentDisposition("inline"))
	if err != nil {
		c.logger.Errorw(ctx, "upload failed", "name", c.name, "err", err)
		return err
	}

	var (
		g     = new(errgroup.Group)
		ch    = make(chan struct{}, 3)
		parts = make(map[int][]byte)
	)

	for i := 0; i < n; i += chunk {
		nn := i + chunk
		nn = min(nn, n)
		parts[part] = body[i:nn]
		part += 1
	}

	for p, b := range parts {
		ch <- struct{}{}
		pp := p
		bb := b

		g.Go(func() error {
			defer func() {
				<-ch
			}()
			var e error
			for i := range 3 {
				e = u.Write(ctx, pp, bb)
				if e != nil {
					c.logger.Infow(ctx, "upload retry", "name", c.name, "uploadId", u.Id(ctx), "part", pp, "total", total, "retry", i, "err", e.Error())
					time.Sleep(time.Millisecond * 50 * (1 << i))
					continue
				}
				break
			}
			c.logger.Infow(ctx, "upload progress", "name", c.name, "uploadId", u.Id(ctx), "part", pp, "total", total, "size", len(bb))
			return e
		})
	}
	if err = g.Wait(); err == nil {
		return nil
	}
	c.logger.Errorw(ctx, "write failed", "name", c.name, "err", err)

	err = c.s3.Clean(ctx, c.name)
	if err != nil {
		c.logger.Errorw(ctx, "clean failed", "name", c.name, "err", err)
		return err
	}

	return nil
}
