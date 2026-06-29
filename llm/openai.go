package llm

import (
	"context"

	"github.com/advancevillage/3rd/logx"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/packages/ssestream"
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
		option.WithBaseURL(opts.baseUrl),
	)
	c.client = &client
	logger.Infow(ctx, "success to create chatgpt client", "sk", opts.sk, "model", opts.model)
	return c, nil
}

func (c *baseGPT) Completion(ctx context.Context, handler StreamHandler, msg []Message) error {
	c.logger.Infow(ctx, "model info", "model", c.opts.model, "sk", c.opts.sk)
	stream := c.client.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
		Messages: toChatMessages(msg),
		Model:    c.opts.model,
	})
	return drainStream(ctx, stream, handler)
}

// toChatMessages 将统一的 Message 转换为 openai 标准对话消息。
func toChatMessages(msg []Message) []openai.ChatCompletionMessageParamUnion {
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
	return chats
}

// drainStream 消费流式响应并驱动 StreamHandler 的生命周期回调。
func drainStream(ctx context.Context, stream *ssestream.Stream[openai.ChatCompletionChunk], handler StreamHandler) error {
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
