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

func Test_new_search_request(t *testing.T) {
	data := map[string]struct {
		req            idx.HybridSearchRequest
		limit          int
		matchThreshold int
		hasFilter      bool
	}{
		"doc-default": {
			req:            idx.NewDocSearchRequest("cards", "苹果"),
			limit:          2,
			matchThreshold: 75,
		},
		"image-default": {
			req:            idx.NewImageSearchRequest("images", "苹果"),
			limit:          2,
			matchThreshold: 75,
		},
		"doc-options": {
			req: idx.NewDocSearchRequest("cards", "苹果",
				idx.WithLimit(10),
				idx.WithMatchThreshold(60),
				idx.WithFilter(idx.Eq("color", "red")),
			),
			limit:          10,
			matchThreshold: 60,
			hasFilter:      true,
		},
	}

	for n, v := range data {
		v := v
		t.Run(n, func(t *testing.T) {
			assert.Equal(t, v.limit, v.req.Limit)
			assert.Equal(t, v.matchThreshold, v.req.MatchThreshold)
			assert.Equal(t, idx.ModeText, v.req.Mode)
			if v.hasFilter {
				assert.NotEmpty(t, v.req.Filter)
			} else {
				assert.Empty(t, v.req.Filter)
			}
		})
	}
}

func Test_hybrid_search(t *testing.T) {
	ctx := context.WithValue(context.TODO(), logx.TraceId, mathx.UUID())
	logger, err := logx.NewLogger("debug")
	assert.Nil(t, err)

	dsn := os.Getenv("IDX_DSN")
	text := "苹果"

	data := map[string]struct {
		req idx.HybridSearchRequest
	}{
		"case-doc": {
			req: idx.NewDocSearchRequest("cards", text),
		},
		"case-image": {
			req: idx.NewImageSearchRequest("images", text),
		},
	}

	for n, v := range data {
		f := func(t *testing.T) {
			c, err := idx.NewHybridSearchClient(ctx, logger, dsn)
			assert.Nil(t, err)

			reply, err := c.HybridSearch(ctx, v.req)
			t.Logf("reply: %+v", reply)
			assert.Nil(t, err)
			assert.NotNil(t, reply)
		}
		t.Run(n, f)
	}
}
