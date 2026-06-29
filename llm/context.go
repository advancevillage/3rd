package llm

import (
	"context"
	"fmt"

	"github.com/advancevillage/3rd/logx"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

var _ LLMContext = &sessionContext{}

type sessionContext struct {
	opts   llmContextOption
	logger logx.ILogger
	client *openai.Client
}

func NewSessionContext(ctx context.Context, logger logx.ILogger, opt ...LLMContextOption) (LLMContext, error) {
	return newSessionContext(ctx, logger, opt...)
}

func newSessionContext(ctx context.Context, logger logx.ILogger, opt ...LLMContextOption) (*sessionContext, error) {
	// 1. 初始化参数
	opts := defaultLLMContextOptions
	for _, o := range opt {
		o.apply(&opts)
	}

	// 2. 火山方舟兼容 openai 标准客户端
	c := &sessionContext{
		opts:   opts,
		logger: logger,
	}
	client := openai.NewClient(
		option.WithAPIKey(opts.sk),
		option.WithBaseURL(opts.baseUrl),
	)
	c.client = &client
	logger.Infow(ctx, "success to create context client", "model", opts.model, "mode", opts.mode)
	return c, nil
}

func (c *sessionContext) Session(ctx context.Context, msg []Message) (string, error) {
	// 请求体对应 POST {baseUrl}/context/create
	req := map[string]any{
		"model":    c.opts.model,
		"mode":     c.opts.mode,
		"messages": toContextMessages(msg),
		"ttl":      c.opts.ttl,
	}
	// 仅 session 模式存在缓存上限, 显式指定 rolling_tokens 截断策略
	if c.opts.mode == "session" && c.opts.maxWindowTokens > 0 {
		req["truncation_strategy"] = map[string]any{
			"type":                  "rolling_tokens",
			"max_window_tokens":     c.opts.maxWindowTokens,
			"rolling_window_tokens": c.opts.rollingWindowTokens,
		}
	}
	// 响应体中 id 即后续对话使用的 context_id
	c.logger.Infow(ctx, "create context cache", "req", req)
	var resp map[string]any
	err := c.client.Post(ctx, "context/create", req, &resp)
	if err != nil {
		c.logger.Errorw(ctx, "fail to create context cache", "err", err, "model", c.opts.model)
		return "", err
	}
	id := fmt.Sprintf("%v", resp["id"])
	c.logger.Infow(ctx, "success to create context cache", "id", id, "mode", c.opts.mode)
	return id, nil
}

func (c *sessionContext) Completion(ctx context.Context, handler StreamHandler, contextId string, msg []Message) error {
	c.logger.Infow(ctx, "context completion", "id", contextId, "model", c.opts.model)
	// 命中 {baseUrl}/context/chat/completions: 在 baseUrl 追加 /context 前缀复用 SDK 流式能力,
	// 并通过 context_id 字段携带缓存 id。
	stream := c.client.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
		Messages: toChatMessages(msg),
		Model:    c.opts.model,
	},
		option.WithBaseURL(c.opts.baseUrl+"/context"),
		option.WithJSONSet("context_id", contextId),
	)
	return drainStream(ctx, stream, handler)
}

// toContextMessages 将统一的 Message 转换为 context/create 的消息体。
func toContextMessages(msg []Message) []map[string]any {
	out := make([]map[string]any, 0, len(msg))
	for i := range msg {
		m := &message{}
		msg[i].apply(m)
		out = append(out, map[string]any{"role": m.role, "content": m.content})
	}
	return out
}
