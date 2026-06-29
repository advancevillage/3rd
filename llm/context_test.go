package llm_test

import (
	"context"
	"os"
	"testing"

	"github.com/advancevillage/3rd/llm"
	"github.com/advancevillage/3rd/logx"
	"github.com/advancevillage/3rd/mathx"
	"github.com/stretchr/testify/assert"
)

func Test_context(t *testing.T) {
	ctx := context.WithValue(context.TODO(), logx.TraceId, mathx.UUID())
	logger, err := logx.NewLogger("debug")
	assert.Nil(t, err)

	c, err := llm.NewSessionContext(ctx, logger,
		llm.WithContextBaseUrl("https://ark.cn-beijing.volces.com/api/v3"),
		llm.WithContextModel("ep-m-20260617230909-pj2tz"),
		llm.WithContextSecret(os.Getenv("DOUBAO_SK")),
		llm.WithContextMode("session"),
		llm.WithContextTTL(3600),
		llm.WithContextMaxWindowTokens(8192),
		llm.WithContextRollingWindowTokens(4096),
	)
	assert.Nil(t, err)

	// 1. 创建上下文缓存，写入初始 system 设定，拿到 context_id
	contextId, err := c.Session(ctx, []llm.Message{
		llm.WithSystemMessage("你是一名平面设计专家，突出创意，与众不同。同时兼顾生产制造的限制。"),
	})
	assert.Nil(t, err)
	assert.NotEmpty(t, contextId)

	// 2. 携带 context_id 进行多轮对话，缓存命中已处理好的 system 设定
	h := &testStreamHandler{t: t}
	err = c.Completion(ctx, h, contextId, []llm.Message{
		llm.WithUserMessage("帮忙生成Q版麻花平面图，1024x1024像素，麦穗形状、纤细，垂直，柔柔的"),
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, h.s)
	assert.Equal(t, 1, h.e)
}
