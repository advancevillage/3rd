package llm

import (
	"context"
	"time"

	"github.com/advancevillage/3rd/logx"
	"github.com/advancevillage/3rd/x"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/packages/ssestream"
	"github.com/openai/openai-go/v3/responses"
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

func (c *baseGPT) Completion(ctx context.Context, handler StreamHandler, msg []Message, opts ...CompletionOption) error {
	c.logger.Infow(ctx, "model info", "model", c.opts.model, "mode", c.opts.mode)
	o := completionOption{}
	for _, x := range opts {
		x.apply(&o)
	}
	switch c.opts.mode {
	case ModeChat:
		return c.streamCompletion(ctx, handler, msg, o)
	default: // ModeResponse
		return c.streamResponse(ctx, handler, msg, o)
	}
}

// streamCompletion 使用 Chat Completion API 流式补全。
// Chat Completion 不承载缓存, 静默忽略 Response 专属选项(web_search/thinking/caching)。
func (c *baseGPT) streamCompletion(ctx context.Context, handler StreamHandler, msg []Message, _ completionOption) error {
	stream := c.client.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
		Messages: toChatMessages(msg),
		Model:    c.opts.model,
	})
	return drainStream(ctx, stream, handler, chatFrame)
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

// streamFrame 单个流事件的解码结果, 是各 API 事件与 StreamHandler 之间的统一中间形态。
type streamFrame struct {
	text  string     // 文本增量, 空串表示该事件无文本输出
	start bool       // 标记流开始(Response API 的 response.created), 触发 OnStart
	end   bool       // 标记流结束(response.completed), opts 携带至 OnEnd
	opts  []x.Option // 该事件携带的生命周期元数据
}

// drainStream 消费流式响应并驱动 StreamHandler 的生命周期回调。
// decode 将单个事件解码为统一的 streamFrame, 由各 API 路径自行实现。
func drainStream[T any](ctx context.Context, stream *ssestream.Stream[T], handler StreamHandler, decode func(T) streamFrame) error {
	defer stream.Close()

	started := false
	var endOpts []x.Option

	for stream.Next() {
		f := decode(stream.Current())
		if !started && (f.start || f.text != "") {
			started = true
			handler.OnStart(ctx, f.opts...)
		}
		if f.text != "" {
			handler.OnChunk(ctx, f.text)
		}
		if f.end {
			endOpts = f.opts
		}
	}

	if err := stream.Err(); err != nil {
		return err
	}

	handler.OnEnd(ctx, endOpts...)
	return nil
}

// chatFrame 从 Chat Completion 流式分片解码 streamFrame。
// Chat Completion 无 created/completed/usage 事件, 仅承载文本, 生命周期元数据为零值。
func chatFrame(chunk openai.ChatCompletionChunk) streamFrame {
	for _, choice := range chunk.Choices {
		if len(choice.Delta.Content) > 0 {
			return streamFrame{text: choice.Delta.Content}
		}
	}
	return streamFrame{}
}

// streamResponse 使用 Response API 流式补全, 按需透传 web_search/thinking/caching/expire_at/previous_response_id。
func (c *baseGPT) streamResponse(ctx context.Context, handler StreamHandler, msg []Message, o completionOption) error {
	params, reqOpts := c.buildResponseParams(msg, o)
	stream := c.client.Responses.NewStreaming(ctx, params, reqOpts...)
	return drainStream(ctx, stream, handler, responseFrame(o))
}

// buildResponseParams 构建 Response API 请求参数与请求级选项, 供流式与非流式共用。
// caching/expire_at 与 thinking 是火山方舟扩展字段, 经 option.WithJSONSet 注入请求体。
// expire_at 是缓存过期的绝对 UTC Unix 时间戳(秒), 由 now+cacheTTL 算出。
func (c *baseGPT) buildResponseParams(msg []Message, o completionOption) (responses.ResponseNewParams, []option.RequestOption) {
	params := responses.ResponseNewParams{
		Model: c.opts.model,
		Input: responses.ResponseNewParamsInputUnion{OfInputItemList: toResponseInput(msg)},
	}
	//-- 火山联网
	//https://www.volcengine.com/docs/82379/1756990?lang=zh
	if o.webSearch {
		params.Tools = []responses.ToolUnionParam{responses.ToolParamOfWebSearch(responses.WebSearchToolTypeWebSearch)}
	}
	if o.prevRespId != "" {
		params.PreviousResponseID = openai.String(o.prevRespId)
	}

	var reqOpts []option.RequestOption
	//-- 火山缓存
	//https://www.volcengine.com/docs/82379/1602228?lang=zh
	if o.caching {
		reqOpts = append(reqOpts, option.WithJSONSet("caching", map[string]any{"type": "enabled"}))
		reqOpts = append(reqOpts, option.WithJSONSet("expire_at", time.Now().Unix()+o.expireAt))
	}
	if o.thinking {
		reqOpts = append(reqOpts, option.WithJSONSet("thinking", map[string]any{"type": "enabled"}))
	}

	return params, reqOpts
}

// responseFrame 构造 Response API 事件的解码器。
// think/cache 取自请求侧参数(o), created 事件仅用于回显; usage 取自 completed 事件。
func responseFrame(o completionOption) func(responses.ResponseStreamEventUnion) streamFrame {
	return func(e responses.ResponseStreamEventUnion) streamFrame {
		switch e.Type {
		case "response.created":
			return streamFrame{start: true, opts: []x.Option{
				x.WithKV(MetaResponseID, e.Response.ID),
				x.WithKV(MetaThinking, o.thinking),
				x.WithKV(MetaCaching, o.caching),
			}}
		case "response.output_text.delta":
			return streamFrame{text: e.Delta}
		case "response.completed":
			u := e.Response.Usage
			return streamFrame{end: true, opts: []x.Option{
				x.WithKV(MetaInputTokens, u.InputTokens),
				x.WithKV(MetaOutputTokens, u.OutputTokens),
				x.WithKV(MetaTotalTokens, u.TotalTokens),
				x.WithKV(MetaCachedTokens, u.InputTokensDetails.CachedTokens),
			}}
		}
		return streamFrame{}
	}
}

// toResponseInput 将统一的 Message 转换为 Response API 的输入条目列表。
func toResponseInput(msg []Message) responses.ResponseInputParam {
	items := make(responses.ResponseInputParam, 0, len(msg))
	for i := range msg {
		m := &message{}
		msg[i].apply(m)
		switch m.role {
		case roleUser:
			items = append(items, responses.ResponseInputItemParamOfMessage(m.content, responses.EasyInputMessageRoleUser))
		case roleAssist:
			items = append(items, responses.ResponseInputItemParamOfMessage(m.content, responses.EasyInputMessageRoleAssistant))
		case roleSystem:
			items = append(items, responses.ResponseInputItemParamOfMessage(m.content, responses.EasyInputMessageRoleSystem))
		}
	}
	return items
}
