package llm

import (
	"context"

	"github.com/advancevillage/3rd/logx"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type LLM interface {
}

var _ LLM = &chatGPT{}

type chatGPT struct {
	opts   llmOption
	logger logx.ILogger
	client *openai.Client
}

func newChatGPT(ctx context.Context, logger logx.ILogger, opt ...LLMOption) (*chatGPT, error) {
	// 1. 初始化参数
	opts := defaultLLMOptions
	for _, o := range opt {
		o.apply(&opts)
	}

	// 2. chatGPT客户端
	client := openai.NewClient(option.WithAPIKey(opts.sk))

	// 3. 返回
	return &chatGPT{
		opts:   opts,
		logger: logger,
		client: client,
	}, nil
}
