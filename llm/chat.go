package llm

import (
	"context"

	"github.com/advancevillage/3rd/logx"
	"github.com/openai/openai-go/v3"
)

var _ LLMChat = &baseGPT{}

func NewBaseChat(ctx context.Context, logger logx.ILogger, opt ...LLMOption) (LLMChat, error) {
	return newBaseGPT(ctx, logger, opt...)
}

// Complete 非流式一次性补全, 返回完整文本; 路由与 Completion 对称。
func (c *baseGPT) Chat(ctx context.Context, msg []Message, opts ...CompletionOption) (string, error) {
	c.logger.Infow(ctx, "model info", "model", c.opts.model, "mode", c.opts.mode)
	o := completionOption{}
	for _, x := range opts {
		x.apply(&o)
	}
	switch c.opts.mode {
	case ModeChat:
		return c.chatComplete(ctx, msg)
	default: // ModeResponse
		return c.responseComplete(ctx, msg, o)
	}
}

// chatComplete 使用 Chat Completion API 非流式补全。
// Chat Completion 不承载缓存, 静默忽略 Response 专属选项(web_search/thinking/caching)。
func (c *baseGPT) chatComplete(ctx context.Context, msg []Message) (string, error) {
	resp, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: toChatMessages(msg),
		Model:    c.opts.model,
	})
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", nil
	}
	return resp.Choices[0].Message.Content, nil
}

// responseComplete 使用 Response API 非流式补全, 与 streamResponse 共用请求参数构建。
func (c *baseGPT) responseComplete(ctx context.Context, msg []Message, o completionOption) (string, error) {
	params, reqOpts := c.buildResponseParams(msg, o)
	resp, err := c.client.Responses.New(ctx, params, reqOpts...)
	if err != nil {
		return "", err
	}
	return resp.OutputText(), nil
}
