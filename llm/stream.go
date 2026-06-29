package llm

import (
	"context"
	"strings"

	"github.com/advancevillage/3rd/logx"
)

type LLMContext interface {
	// Session 创建上下文缓存(context/create), 返回 context_id
	Session(ctx context.Context, msg []Message) (string, error)
	// Completion 携带 context_id 进行流式对话(context/chat/completions)
	Completion(ctx context.Context, handler StreamHandler, contextId string, msg []Message) error
}

type LLMStream interface {
	Completion(ctx context.Context, handler StreamHandler, msg []Message) error
}

type StreamHandler interface {
	OnStart(ctx context.Context)
	OnChunk(ctx context.Context, chunk string)
	OnEnd(ctx context.Context)
}

var _ StreamHandler = &emptyStreamHandler{}

type emptyStreamHandler struct{}

func (emptyStreamHandler) OnStart(ctx context.Context)               {}
func (emptyStreamHandler) OnChunk(ctx context.Context, chunk string) {}
func (emptyStreamHandler) OnEnd(ctx context.Context)                 {}

// 缓存流事件处理
var _ StreamHandler = &bufferStreamHandler{}

type bufferStreamHandler struct {
	buf     string
	logger  logx.ILogger
	handler StreamHandler
}

func (h *bufferStreamHandler) OnStart(ctx context.Context) {
	if len(h.buf) == 0 {
		h.handler.OnStart(ctx)
	}
}

func (h *bufferStreamHandler) OnEnd(ctx context.Context) {
	if len(h.buf) > 0 {
		h.handler.OnChunk(ctx, h.buf)
		h.buf = ""
	}
	h.handler.OnEnd(ctx)
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
