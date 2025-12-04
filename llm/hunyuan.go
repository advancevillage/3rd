package llm

import (
	"context"
	"encoding/json"

	"github.com/advancevillage/3rd/logx"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	hunyuan "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/hunyuan/v20230901"
)

type LLMStream interface {
	Completion(ctx context.Context, msg []Message) error
}

var _ LLMStream = &hunYuan{}

// HunYuan事件结构
type hunYuanEvent struct {
	Choices []struct {
		Index        int    `json:"Index"`
		FinishReason string `json:"FinishReason"`
		Delta        struct {
			Role             string `json:"Role"`
			Content          string `json:"Content"`
			ReasoningContent string `json:"ReasoningContent"`
		} `json:"Delta"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"PromptTokens"`
		CompletionTokens int `json:"CompletionTokens"`
		TotalTokens      int `json:"TotalTokens"`
	}
}

// 官方文档
// https://cloud.tencent.com/document/product/1729/111007
type hunYuan struct {
	opts   llmStreamOption
	logger logx.ILogger
	hyc    *hunyuan.Client
}

func NewHunYuan(ctx context.Context, logger logx.ILogger, opt ...LLMStreamOption) (LLMStream, error) {
	return newHunYuan(ctx, logger, opt...)
}

func newHunYuan(ctx context.Context, logger logx.ILogger, opt ...LLMStreamOption) (*hunYuan, error) {
	// 1. 初始化参数
	opts := defaultLLMStreamOptions
	for _, o := range opt {
		o.apply(&opts)
	}

	// 2. hunYuan客户端(openai兼容)
	c := &hunYuan{
		opts:   opts,
		logger: logger,
	}
	// 3. HunYuan SDK
	credential := common.NewCredential(opts.ak, opts.sk)
	cpf := profile.NewClientProfile()
	// 流式接口耗时较长
	cpf.HttpProfile.ReqTimeout = opts.timeout
	hyc, err := hunyuan.NewClient(credential, opts.region, cpf)
	if err != nil {
		logger.Errorw(ctx, "failed to create hunYuan client", "err", err)
		return nil, err
	}
	c.hyc = hyc
	return c, nil
}

func (c *hunYuan) Completion(ctx context.Context, msg []Message) error {
	return c.stream(ctx, msg)
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
	var chunk = &hunYuanEvent{}
	var first = true
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
		//s.logger.Infow(ctx, "stream event", "chunk", string(evt.Data))
		switch {
		case first:
			s.opts.handler.OnStart(ctx)
			s.opts.handler.OnChunk(ctx, chunk.Choices[0].Delta.Content)
			first = false

		case len(chunk.Choices[0].FinishReason) > 0:
			s.opts.handler.OnChunk(ctx, chunk.Choices[0].Delta.Content)
			s.opts.handler.OnEnd(ctx)

		default:
			s.opts.handler.OnChunk(ctx, chunk.Choices[0].Delta.Content)
		}
	}
	return nil
}
