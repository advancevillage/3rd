package llm

import (
	"context"
	"strings"

	"github.com/advancevillage/3rd/logx"
	"github.com/advancevillage/3rd/x"
)

type LLMStream interface {
	Completion(ctx context.Context, handler StreamHandler, msg []Message, opts ...CompletionOption) error
}

type StreamHandler interface {
	OnStart(ctx context.Context, opts ...x.Option)
	OnChunk(ctx context.Context, chunk string)
	OnEnd(ctx context.Context, opts ...x.Option)
}

// 流式元数据 key: 经 x.WithKV 写入 OnStart/OnEnd 的 x.Option, 由 x.Builder 读取。
// OnStart 阶段携带 response id 与 think/cache(请求侧); OnEnd 阶段携带 usage。
const (
	MetaResponseID   = "response_id"   // string, OnStart
	MetaThinking     = "thinking"      // bool, OnStart
	MetaCaching      = "caching"       // bool, OnStart
	MetaInputTokens  = "input_tokens"  // int64, OnEnd
	MetaOutputTokens = "output_tokens" // int64, OnEnd
	MetaTotalTokens  = "total_tokens"  // int64, OnEnd
	MetaCachedTokens = "cached_tokens" // int64, OnEnd: 输入缓存命中
)

var _ StreamHandler = &emptyStreamHandler{}

type emptyStreamHandler struct{}

func (emptyStreamHandler) OnStart(ctx context.Context, opts ...x.Option) {}
func (emptyStreamHandler) OnChunk(ctx context.Context, chunk string)     {}
func (emptyStreamHandler) OnEnd(ctx context.Context, opts ...x.Option)   {}

// 缓存流事件处理
var _ StreamHandler = &bufferStreamHandler{}

type bufferStreamHandler struct {
	buf     string
	logger  logx.ILogger
	handler StreamHandler
}

func (h *bufferStreamHandler) OnStart(ctx context.Context, opts ...x.Option) {
	if len(h.buf) == 0 {
		h.handler.OnStart(ctx, opts...)
	}
}

func (h *bufferStreamHandler) OnEnd(ctx context.Context, opts ...x.Option) {
	if len(h.buf) > 0 {
		h.handler.OnChunk(ctx, h.buf)
		h.buf = ""
	}
	h.handler.OnEnd(ctx, opts...)
}

func (h *bufferStreamHandler) OnChunk(ctx context.Context, chunk string) {
	h.buf += chunk
	runes := []rune(h.buf)
	lastSep := -1
	for i, r := range runes {
		switch r {
		case '。', '！', '？', '；', '.', '!', '?', ';', '\n':
			lastSep = i
		}
	}
	if lastSep < 0 {
		return
	}

	h.handler.OnChunk(ctx, strings.ReplaceAll(string(runes[:lastSep+1]), "\n", `\n`))
	h.buf = string(runes[lastSep+1:])
}

func NewBufferStreamHandler(ctx context.Context, logger logx.ILogger, handler StreamHandler) StreamHandler {
	return &bufferStreamHandler{handler: handler, logger: logger}
}
