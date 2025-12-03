package llm

import (
	"context"
	"encoding/json"

	"github.com/advancevillage/3rd/logx"
	"github.com/advancevillage/3rd/x"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/regions"
	hunyuan "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/hunyuan/v20230901"
)

var _ LLM = &hunYuan{}

// 官方文档
// https://cloud.tencent.com/document/product/1729/111007
type hunYuan struct {
	opts   llmOption
	logger logx.ILogger
	client *openai.Client
	hyc    *hunyuan.Client
}

func NewHunYuan(ctx context.Context, logger logx.ILogger, opt ...LLMOption) (LLM, error) {
	return newHunYuan(ctx, logger, opt...)
}

func newHunYuan(ctx context.Context, logger logx.ILogger, opt ...LLMOption) (*hunYuan, error) {
	// 1. 初始化参数
	opts := defaultLLMOptions
	for _, o := range opt {
		o.apply(&opts)
	}

	// 2. hunYuan客户端(openai兼容)
	c := &hunYuan{
		opts:   opts,
		logger: logger,
	}
	client := openai.NewClient(option.WithAPIKey(opts.sk), option.WithBaseURL("https://api.hunYuan.cloud.tencent.com/v1/"))
	c.client = &client
	logger.Infow(ctx, "success to crate huanyuan client", "sk", opts.sk, "model", opts.model)

	// 3. HunYuan SDK
	credential := common.NewCredential(opts.ak1, opts.sk1)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.ReqTimeout = 400 // 流式接口耗时较长
	hyc, err := hunyuan.NewClient(credential, regions.Guangzhou, cpf)
	if err != nil {
		logger.Errorw(ctx, "failed to create hunYuan client", "err", err)
		return nil, err
	}
	c.hyc = hyc
	return c, nil
}

func (c *hunYuan) Completion(ctx context.Context, msg []Message, schema x.Builder, resp any) error {
	// 1. 构建消息列表
	// 2. 多轮对话
	//return c.complete(ctx, chats, schema.Build(), resp)
	return c.stream(ctx, msg)
}

func (c *hunYuan) complete(ctx context.Context, chats []openai.ChatCompletionMessageParamUnion, schema any, resp any) error {
	// 1. 结构化
	reply, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model:    c.opts.model,
		Messages: chats,
		N:        openai.Int(1),
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
		c.logger.Errorw(ctx, "failed to create hunYuan intent", "err", err)
		return err
	}
	for i := range reply.Choices {
		c.logger.Infow(ctx, "hunYuan intent reply", "content", reply.Choices[i].Message.Content)
		data := []byte(reply.Choices[i].Message.Content)
		err := json.Unmarshal(data, resp)
		if err != nil {
			c.logger.Warnw(ctx, "failed to unmarshal hunYuan intent", "err", err)
			continue
		}
		break
	}
	c.logger.Infow(ctx, "success to create hunYuan intent", "reply", resp)
	return nil
}

func (s *hunYuan) stream(ctx context.Context, msg []Message) error {
	// 1. 构建混元模型请求
	req := hunyuan.NewChatCompletionsRequest()
	for i := range msg {
		m := &message{}
		msg[i].apply(m)
		req.Messages = append(req.Messages, &hunyuan.Message{
			Role:    common.StringPtr(m.role),
			Content: common.StringPtr(m.content),
		})
	}
	req.Stream = common.BoolPtr(true)
	req.Model = common.StringPtr(s.opts.model)
	// 2. 请求模型
	reply, err := s.hyc.ChatCompletionsWithContext(ctx, req)
	if err != nil {
		s.logger.Errorw(ctx, "failed to create hunYuan std stream", "err", err)
		return err
	}

	// 3. SSE返回
	var chunk = &HunYuanEvent{}
	for evt := range reply.Events {
		err := json.Unmarshal(evt.Data, chunk)
		if err != nil {
			s.logger.Errorw(ctx, "failed to unmarshal hunYuan stream event", "err", err, "data", string(evt.Data))
			continue
		}
		if len(chunk.Choices) <= 0 {
			s.logger.Infow(ctx, "hunYuan stream event no choices", "data", string(evt.Data))
			continue
		}
		if len(chunk.Choices[0].Delta.FinishReason) > 0 {
			s.opts.sf(ctx, "STOP")
		} else {
			s.opts.sf(ctx, chunk.Choices[0].Delta.Content)
		}
		//s.logger.Infow(ctx, "hunYuan stream chunk total tokens", "tokens", chunk.Usage.TotalTokens)
	}
	return nil
}
