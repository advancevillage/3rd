package llm

import (
	"context"

	"github.com/openai/openai-go/v3"
)

type StreamFunc func(ctx context.Context, chunk string)

type LLMOption interface {
	apply(*llmOption)
}

func WitChatGPTSecret(sk string) LLMOption {
	return newFuncLLMOption(func(o *llmOption) {
		o.sk = sk
	})
}

func WithO3MiniModel() LLMOption {
	return newFuncLLMOption(func(o *llmOption) {
		o.model = openai.ChatModelO3Mini
	})
}

func With4OModel() LLMOption {
	return newFuncLLMOption(func(o *llmOption) {
		o.model = openai.ChatModelGPT4o
	})
}

func WithModel(model string) LLMOption {
	return newFuncLLMOption(func(o *llmOption) {
		o.model = model
	})
}

func WithChatGPTProxy(proxy string) LLMOption {
	return newFuncLLMOption(func(o *llmOption) {
		o.proxy = proxy
	})
}

func WithSecret(ak string, sk string) LLMOption {
	return newFuncLLMOption(func(o *llmOption) {
		o.ak1 = ak
		o.sk1 = sk
	})
}

func WithStreamFunc(sf StreamFunc) LLMOption {
	return newFuncLLMOption(func(o *llmOption) {
		o.sf = sf
	})
}

type llmOption struct {
	sk    string
	ak1   string
	sk1   string
	sf    StreamFunc
	proxy string
	model string
}

var defaultLLMOptions = llmOption{
	sf:    func(ctx context.Context, chunk string) {},
	model: openai.ChatModelGPT4oMini,
}

type funcLLMOption struct {
	f func(*llmOption)
}

func (fdo *funcLLMOption) apply(do *llmOption) {
	fdo.f(do)
}

func newFuncLLMOption(f func(*llmOption)) *funcLLMOption {
	return &funcLLMOption{
		f: f,
	}
}

// 消息优先级: system > user > assistant(历史)
type Message interface {
	apply(*message)
}

// user — 用户的真实需求（普通优先级）
func WithUserMessage(content string) Message {
	return newFuncMessage(func(m *message) {
		m.role = "user"
		m.content = content
	})
}

// system — 角色设定、全局规则（最高级别）
func WithSystemMessage(content string) Message {
	return newFuncMessage(func(m *message) {
		m.role = "system"
		m.content = content
	})
}

// assistant — 模型输出
func WithassistantMessage(content string) Message {
	return newFuncMessage(func(m *message) {
		m.role = "assistant"
		m.content = content
	})
}

type message struct {
	role    string
	content string
}

type funcMessage struct {
	f func(*message)
}

func (fdo *funcMessage) apply(do *message) {
	fdo.f(do)
}

func newFuncMessage(f func(*message)) *funcMessage {
	return &funcMessage{
		f: f,
	}
}
