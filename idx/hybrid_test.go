package idx_test

import (
	"context"
	"os"
	"testing"

	"github.com/advancevillage/3rd/idx"
	"github.com/advancevillage/3rd/logx"
	"github.com/advancevillage/3rd/mathx"
	"github.com/stretchr/testify/assert"
)

func Test_hybrid_search(t *testing.T) {
	ctx := context.WithValue(context.TODO(), logx.TraceId, mathx.UUID())
	logger, err := logx.NewLogger("debug")
	assert.Nil(t, err)

	dsn := os.Getenv("IDX_DSN")
	dataset := os.Getenv("IDX_DATASET")
	text := os.Getenv("IDX_SEARCH_TEXT")
	if dsn == "" || dataset == "" {
		t.Skip("IDX_DSN and IDX_DATASET are required")
	}
	if text == "" {
		text = "蓝色汽车"
	}

	data := map[string]struct {
		req idx.HybridSearchRequest
	}{
		"case-doc": {
			req: idx.NewDocSearchRequest(dataset, text),
		},
		"case-image": {
			req: idx.NewImageSearchRequest(dataset, text),
		},
	}

	for n, v := range data {
		f := func(t *testing.T) {
			c, err := idx.NewHybridSearchClient(ctx, logger, dsn)
			assert.Nil(t, err)

			reply, err := c.HybridSearch(ctx, v.req)
			assert.Nil(t, err)
			assert.NotNil(t, reply)
			assert.NotEmpty(t, reply.RequestId)
			t.Logf("request_id: %s", reply.RequestId)
		}
		t.Run(n, f)
	}
}
