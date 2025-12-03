package llm

import (
	"context"

	"github.com/advancevillage/3rd/logx"
)

// HunYuan事件结构
type HunYuanEvent struct {
	Choices []struct {
		Index int `json:"Index"`
		Delta struct {
			Role             string `json:"Role"`
			Content          string `json:"Content"`
			ReasoningContent string `json:"ReasoningContent"`
			FinishReason     string `json:"FinishReason"`
		} `json:"Delta"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"PromptTokens"`
		CompletionTokens int `json:"CompletionTokens"`
		TotalTokens      int `json:"TotalTokens"`
	}
}

// 缓存流事件处理
type bufferEvent struct {
	opts llmOption
	buf  string
}

func (b *bufferEvent) Process(ctx context.Context, chunk string) {
	// 1. 拼接上下文
	b.buf += chunk
	var (
		runes    = []rune(b.buf)
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
			offset = i + 1
			exitLoop = true
		}
	}
	if offset >= 0 {
		b.buf = string(runes[offset:])
		b.opts.sf(ctx, string(runes[:offset]))
	}
}

func NewBufferEvent(ctx context.Context, logger logx.ILogger, opt ...LLMOption) StreamFunc {
	// 1. 初始化参数
	opts := defaultLLMOptions
	for _, o := range opt {
		o.apply(&opts)
	}
	be := &bufferEvent{opts: opts}
	return be.Process
}
