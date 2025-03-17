package llm

import (
	"context"
	"encoding/json"

	"github.com/advancevillage/3rd/logx"
	"github.com/advancevillage/3rd/x"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type LLM interface {
	Completion(ctx context.Context, role string, query string, schema x.Builder, resp any) error
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
	client := openai.NewClient(option.WithAPIKey(opts.sk))
	logger.Infow(ctx, "success to crate chatgpt client", "opts", opts)

	// 3. 返回
	return &chatGPT{
		opts:   opts,
		logger: logger,
		client: client,
	}, nil
}

// 使用ChatGPT生成一次性回复
func (c *chatGPT) Completion(ctx context.Context, role string, query string, schema x.Builder, resp any) error {
	return c.complete(ctx, role, query, schema.Build(), resp)
}

func (c *chatGPT) complete(ctx context.Context, role any, query any, schema any, resp any) error {
	// 1. 结构化
	reply, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: openai.F(c.opts.model),
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.ChatCompletionMessageParam{
				Role:    openai.F(openai.ChatCompletionMessageParamRoleUser),
				Content: openai.F(query),
			},
			openai.ChatCompletionMessageParam{
				Role:    openai.F(openai.ChatCompletionMessageParamRoleSystem),
				Content: openai.F(role),
			},
		}),
		ResponseFormat: openai.F[openai.ChatCompletionNewParamsResponseFormatUnion](
			openai.ResponseFormatJSONSchemaParam{
				Type: openai.F(openai.ResponseFormatJSONSchemaTypeJSONSchema),
				JSONSchema: openai.F(openai.ResponseFormatJSONSchemaJSONSchemaParam{
					Name:        openai.F("schema"),
					Description: openai.F("The schema of the response"),
					Schema:      openai.F(schema),
					Strict:      openai.Bool(true),
				}),
			},
		),
	})
	if err != nil {
		c.logger.Errorw(ctx, "failed to create chatgpt intent", "err", err, "query", query)
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
