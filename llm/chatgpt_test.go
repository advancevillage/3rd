package llm_test

import (
	"context"
	"testing"

	"github.com/advancevillage/3rd/llm"
	"github.com/advancevillage/3rd/logx"
	"github.com/advancevillage/3rd/mathx"
	"github.com/stretchr/testify/assert"
)

func Test_completion(t *testing.T) {
	ctx := context.WithValue(context.TODO(), logx.TraceId, mathx.UUID())
	logger, err := logx.NewLogger("debug")
	assert.Nil(t, err)

	cli, err := llm.NewChatGPT(ctx, logger)
	assert.Nil(t, err)

	t.Log(cli)
}
