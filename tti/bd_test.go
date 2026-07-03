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

func Test_bd_imagen(t *testing.T) {
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
			prompt: "星际穿越，黑洞，黑洞里冲出一辆快支离破碎的复古列车，抢视觉冲击力，电影大片，末日既视感，动感，对比色，oc渲染，光线追踪，动态模糊，景深，超现实主义，深蓝，画面通过细腻的丰富的色彩层次塑造主体与场景，质感真实，暗黑风背景的光影效果营造出氛围，整体兼具艺术幻想感，夸张的广角透视效果，耀光，反射，极致的光影，强引力，吞噬",
		},
	}

	for n, v := range data {
		f := func(t *testing.T) {
			c, err := tti.NewBdImageClient(ctx, logger, s3, tti.WitGenerateSecret(os.Getenv("DOUBAO_SK")), tti.WithGenerateModel("doubao-seedream-5-0-260128"))
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
