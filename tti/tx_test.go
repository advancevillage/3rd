package tti_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/advancevillage/3rd/dbx"
	"github.com/advancevillage/3rd/logx"
	"github.com/advancevillage/3rd/mathx"
	"github.com/advancevillage/xmagic/internal/tti"
	"github.com/stretchr/testify/assert"
)

func Test_stability(t *testing.T) {
	ctx := context.WithValue(context.TODO(), logx.TraceId, mathx.UUID())
	ctx = context.WithValue(ctx, "accountId", "AC1905654379034595328")
	logger, err := logx.NewLogger("debug")
	assert.Nil(t, err)
	s3, err := dbx.NewCosClient(ctx, os.Getenv("COS_DSN"))
	assert.Nil(t, err)

	var data = map[string]struct {
		prompt string
	}{
		"case1": {
			prompt: "A post-war cityscape with towering skyscrapers, partially damaged but under reconstruction. The buildings showcase a mix of futuristic and brutalist architecture, with scaffolding and cranes symbolizing recovery. Some structures have overgrown vegetation, while others feature large banners of peace. The atmosphere is melancholic yet hopeful, with warm sunlight breaking through scattered clouds, casting long shadows over the ruined yet resilient urban landscape.",
		},
	}

	for n, v := range data {
		f := func(t *testing.T) {
			c, err := tti.NewStabilityClient(ctx, logger, s3, tti.WitGenerateSecret(os.Getenv("STABILITY_SK")))
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

					err = s3.Clean(ctx, name)
					assert.Nil(t, err)
					return
				}
			}
		}
		t.Run(n, f)
	}
}
