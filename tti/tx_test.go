package tti_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/advancevillage/3rd/dbx"
	"github.com/advancevillage/3rd/logx"
	"github.com/advancevillage/3rd/mathx"
	"github.com/advancevillage/3rd/tti"
	"github.com/stretchr/testify/assert"
)

func Test_tx_imagen(t *testing.T) {
	ctx := context.WithValue(context.TODO(), logx.TraceId, mathx.UUID())
	ctx = context.WithValue(ctx, "accountId", "AC1905654379034595328")
	logger, err := logx.NewLogger("debug")
	assert.Nil(t, err)
	var (
		ak     = os.Getenv("COS_AK")
		sk     = os.Getenv("COS_SK")
		domain = os.Getenv("COS_DOMAIN")
	)
	s3, err := dbx.NewCosS3(ctx, "xmagic-1259635961", "accelerate", ak, sk, domain)
	assert.Nil(t, err)

	data := map[string]struct {
		prompt string
	}{
		"case1": {
			prompt: "一只在草地上奔跑的金毛犬，阳光明媚，背景是蓝天白云",
		},
	}

	for n, v := range data {
		f := func(t *testing.T) {
			c, err := tti.NewTxImageClient(ctx, logger, s3, tti.WitGenerateSecret(os.Getenv("TX_SK")), tti.WithGenerateModel("hy-image-v3.0"))
			assert.NoError(t, err)

			d, err := c.Generate(ctx, v.prompt)
			assert.Nil(t, err)

			tricker := time.NewTicker(2 * time.Second)
			defer tricker.Stop()

			for {
				select {
				case <-ctx.Done():
					t.Fatal("generate timeout 2 minutes")
					return

				case <-tricker.C:
					p, err := d.Progress(ctx)
					if err != nil {
						t.Fatal(err)
						break
					}
					if p < 100 {
						break
					}
					name, err := d.Name(ctx)
					assert.Nil(t, err)
					assert.NotEmpty(t, name)

					exist, err := s3.Exist(ctx, name)
					assert.Nil(t, err)
					assert.Equal(t, true, exist)

					uri, err := d.Resource(ctx)
					assert.Nil(t, err)
					assert.NotEmpty(t, uri)
					t.Log(uri)

					// err = s3.Clean(ctx, name)
					assert.Nil(t, err)
					return
				}
			}
		}
		t.Run(n, f)
	}
}
