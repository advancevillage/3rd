package llm

import (
	"context"

	"github.com/advancevillage/3rd/logx"
)

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
	// 1. 拼接上下文
	h.buf += chunk
	var (
		runes = []rune(h.buf)
		start = 0
	)
	// 2. 处理分隔符
	for i := range runes {
		switch runes[i] {
		case '。', '！', '？', '；', '.', '!', '?', ';':
			// Emit chunk up to and including separator
			h.handler.OnChunk(ctx, string(runes[start:i+1]))
			start = i + 1
		case '\n': // 换行符
			runes[i] = ' '
			h.handler.OnChunk(ctx, string(runes[start:i+1]))
			start = i + 1
		}
	}

	// 3. 更新缓冲区
	if start < len(runes) {
		h.buf = string(runes[start:])
	} else {
		h.buf = ""
	}
}

func NewBufferStreamHandler(ctx context.Context, logger logx.ILogger, handler StreamHandler) StreamHandler {
	logger.Infow(ctx, "buffer stream handler created", "handler", handler)
	return &bufferStreamHandler{handler: handler, logger: logger}
}
