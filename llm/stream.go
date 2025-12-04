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
	opts llmStreamOption
	buf  string
}

func (h *bufferStreamHandler) OnStart(ctx context.Context) {
	if len(h.buf) <= 0 {
		h.opts.handler.OnStart(ctx)
	}
}

func (h *bufferStreamHandler) OnEnd(ctx context.Context) {
	if len(h.buf) > 0 {
		h.opts.handler.OnChunk(ctx, h.buf)
		h.buf = ""
	}
	h.opts.handler.OnEnd(ctx)
}

func (h *bufferStreamHandler) OnChunk(ctx context.Context, chunk string) {
	// 1. 拼接上下文
	h.buf += chunk
	var (
		runes    = []rune(h.buf)
		offset   = -1
		exitLoop = false
	)
	// 2. 处理分隔符
	for i := 0; i < len(runes) && !exitLoop; i++ {
		switch runes[i] {
		case '。', '！', '？', '；':
			offset = i + 1
		case '.', '!', '?', ';':
			offset = i + 1
		case '\n': // 换行符
			runes[i] = ' '
			offset = i + 1
			exitLoop = true
		}
	}
	// 3. 触发事件
	if offset >= 0 {
		h.buf = string(runes[offset:])
		h.opts.handler.OnChunk(ctx, string(runes[:offset]))
	}
}

func NewBufferStreamHandler(ctx context.Context, logger logx.ILogger, opt ...LLMStreamOption) StreamHandler {
	// 1. 初始化参数
	opts := defaultLLMStreamOptions
	for _, o := range opt {
		o.apply(&opts)
	}
	return &bufferStreamHandler{opts: opts}
}
