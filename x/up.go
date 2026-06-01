package x

import (
	"context"
	"time"

	"github.com/advancevillage/3rd/dbx"
	"github.com/advancevillage/3rd/logx"
	"golang.org/x/sync/errgroup"
)

func Upload(ctx context.Context, logger logx.ILogger, s3 dbx.S3, name string, body []byte) error {
	var (
		n     = len(body)
		part  = 0
		chunk = 1 << 20
		total = n / chunk
	)
	if n%chunk > 0 {
		total += 1
	}
	logger.Infow(ctx, "upload info", "name", name, "size", n, "total", total, "chunk", chunk)

	u, err := s3.MultiUpload(ctx, name, total)
	if err != nil {
		logger.Errorw(ctx, "upload failed", "name", name, "err", err)
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
					logger.Infow(ctx, "upload retry", "name", name, "uploadId", u.Id(ctx), "part", pp, "total", total, "retry", i, "err", e.Error())
					time.Sleep(time.Millisecond * 50 * (1 << i))
					continue
				}
				break
			}
			logger.Infow(ctx, "upload progress", "name", name, "uploadId", u.Id(ctx), "part", pp, "total", total, "size", len(bb))
			return e
		})
	}
	if err = g.Wait(); err == nil {
		return nil
	}
	logger.Errorw(ctx, "write failed", "name", name, "err", err)

	err = s3.Clean(ctx, name)
	if err != nil {
		logger.Errorw(ctx, "clean failed", "name", name, "err", err)
		return err
	}

	return nil
}
