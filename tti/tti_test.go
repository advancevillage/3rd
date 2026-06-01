package tti

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/advancevillage/3rd/dbx"
	"github.com/advancevillage/3rd/logx"
	"github.com/advancevillage/3rd/mathx"
	"github.com/advancevillage/3rd/netx"
	"github.com/advancevillage/3rd/x"
	"github.com/stretchr/testify/assert"
)

func Test_upload_stability(t *testing.T) {
	ctx := context.WithValue(context.TODO(), logx.TraceId, mathx.UUID())
	logger, err := logx.NewLogger("debug")
	assert.Nil(t, err)
	s3, err := dbx.NewCosClient(ctx, os.Getenv("COS_DSN"))
	assert.Nil(t, err)
	ider, err := mathx.NewSnowFlake(678)
	assert.Nil(t, err)

	transport, err := netx.NewHttpClient(context.TODO(), logger,
		netx.WithClientTimeout(600),
	)
	if err != nil {
		t.Fatal(err)
		return
	}

	data := map[string]struct {
		name string
	}{
		"case1": {
			name: "t/img_77c2e1edcfa8a19baed78a8472ac627f8cdcb781d7ee76777007d46ac8f3471f.png",
		},
	}

	for n, v := range data {
		f := func(t *testing.T) {
			uri, err := s3.Url(ctx, v.name)
			assert.NoError(t, err)
			t.Log(uri)

			res, err := transport.Get(ctx, uri, x.NewBuilder(), x.NewBuilder())
			assert.Nil(t, err)
			d, err := newDescriptor(ctx, logger, s3, ider, WithGeneratePrefix(fmt.Sprintf("test/%s", STABILITY)))
			assert.Nil(t, err)

			d.reply <- res

			for d.progress < 100 {
			}

			exist, err := s3.Exist(ctx, d.name)
			assert.Nil(t, d.err)
			assert.Equal(t, true, exist)

			err = s3.Clean(ctx, d.name)
			assert.Nil(t, err)
		}
		t.Run(n, f)
	}
}
