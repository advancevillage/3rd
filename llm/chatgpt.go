package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/advancevillage/3rd/logx"
	"github.com/advancevillage/3rd/x"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

type LLM interface {
	Completion(ctx context.Context, msg []Message, schema x.Builder, resp any) error
}

var _ LLM = &chatGPT{}

// 官方文档
// https://pkg.go.dev/github.com/openai/openai-go#section-readme
type chatGPT struct {
	opts   llmOption
	logger logx.ILogger
	client *openai.Client
}

func NewChatGPT(ctx context.Context, logger logx.ILogger, opt ...LLMOption) (LLM, error) {
	return newChatGPT(ctx, logger, opt...)
}

func newChatGPT(ctx context.Context, logger logx.ILogger, opt ...LLMOption) (*chatGPT, error) {
	// 1. 初始化参数
	opts := defaultLLMOptions
	for _, o := range opt {
		o.apply(&opts)
	}

	// 2. chatGPT客户端
	c := &chatGPT{
		opts:   opts,
		logger: logger,
	}
	client := openai.NewClient(option.WithAPIKey(opts.sk), option.WithHTTPClient(c.buildClient(ctx)))
	c.client = &client
	logger.Infow(ctx, "success to crate chatgpt client", "sk", opts.sk, "model", opts.model)
	return c, nil
}

func (c *chatGPT) Completion(ctx context.Context, msg []Message, schema x.Builder, resp any) error {
	// 1. 构建消息列表
	chats := make([]openai.ChatCompletionMessageParamUnion, 0, len(msg))
	for i := range msg {
		m := &message{}
		msg[i].apply(m)
		switch m.role {
		case "user":
			chats = append(chats, openai.UserMessage(m.content))
		case "assistant":
			chats = append(chats, openai.AssistantMessage(m.content))
		case "system":
			chats = append(chats, openai.SystemMessage(m.content))
		}
	}
	// 2. 多轮对话
	return c.complete(ctx, chats, schema.Build(), resp)
}

func (c *chatGPT) complete(ctx context.Context, chats []openai.ChatCompletionMessageParamUnion, schema any, resp any) error {
	// 1. 结构化
	reply, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model:    c.opts.model,
		Messages: chats,
		ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONSchema: &openai.ResponseFormatJSONSchemaParam{
				JSONSchema: openai.ResponseFormatJSONSchemaJSONSchemaParam{
					Name:        "schema",
					Description: openai.String("The schema of the response"),
					Schema:      schema,
					Strict:      openai.Bool(true),
				},
			},
		},
	})
	if err != nil {
		c.logger.Errorw(ctx, "failed to create chatgpt intent", "err", err)
		return err
	}
	for i := range reply.Choices {
		data := []byte(reply.Choices[i].Message.Content)
		err := json.Unmarshal(data, resp)
		if err != nil {
			c.logger.Warnw(ctx, "failed to unmarshal chatgpt intent", "err", err)
			continue
		}
		break
	}
	c.logger.Infow(ctx, "success to create chatgpt intent", "reply", resp)
	return nil
}

func (c *chatGPT) buildClient(ctx context.Context) *http.Client {
	// 1. 自定义httpClient
	client := &http.Client{}
	// 2. 构建代理
	proxy, err := url.Parse(c.opts.proxy)
	if err != nil {
		c.logger.Errorw(ctx, "parse proxy url error", "error", err, "proxy", c.opts.proxy)
		return client
	}
	if proxy.Scheme != "http" {
		return client
	}
	client.Transport = &http.Transport{Proxy: http.ProxyURL(proxy)}
	// 3. 返回
	return client
}
