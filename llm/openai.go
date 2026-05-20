package llm

import (
	"context"

	"github.com/advancevillage/3rd/logx"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

var _ LLMStream = &baseGPT{}

type baseGPT struct {
	opts   llmOption
	logger logx.ILogger
	client *openai.Client
}

func NewBaseGPT(ctx context.Context, logger logx.ILogger, opt ...LLMOption) (LLMStream, error) {
	return newBaseGPT(ctx, logger, opt...)
}

func newBaseGPT(ctx context.Context, logger logx.ILogger, opt ...LLMOption) (*baseGPT, error) {
	// 1. 初始化参数
	opts := defaultLLMOptions
	for _, o := range opt {
		o.apply(&opts)
	}

	// 2. Openai标准客户端
	c := &baseGPT{
		opts:   opts,
		logger: logger,
	}
	client := openai.NewClient(
		option.WithAPIKey(opts.sk),
		option.WithBaseURL(opts.host),
	)
	c.client = &client
	logger.Infow(ctx, "success to create chatgpt client", "sk", opts.sk, "model", opts.model)
	return c, nil
}

func (c *baseGPT) Completion(ctx context.Context, handler StreamHandler, msg []Message) error {
	chats := make([]openai.ChatCompletionMessageParamUnion, 0, len(msg))
	for i := range msg {
		m := &message{}
		msg[i].apply(m)
		switch m.role {
		case roleUser:
			chats = append(chats, openai.UserMessage(m.content))
		case roleAssist:
			chats = append(chats, openai.AssistantMessage(m.content))
		case roleSystem:
			chats = append(chats, openai.SystemMessage(m.content))
		}
	}
	return c.complete(ctx, chats, handler)
}

func (c *baseGPT) complete(ctx context.Context, chats []openai.ChatCompletionMessageParamUnion, handler StreamHandler) error {
	stream := c.client.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
		Messages: chats,
		Model:    c.opts.model,
	})
	defer stream.Close()

	first := true

	for stream.Next() {
		chunk := stream.Current()
		if first {
			first = false
			handler.OnStart(ctx)
		}
		for _, choice := range chunk.Choices {
			if len(choice.Delta.Content) > 0 {
				handler.OnChunk(ctx, choice.Delta.Content)
			}
		}
	}

	if err := stream.Err(); err != nil {
		return err
	}

	handler.OnEnd(ctx)
	return nil
}
