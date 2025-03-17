package llm

import "github.com/openai/openai-go"

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

type llmOption struct {
	sk    string
	model string
}

var defaultLLMOptions = llmOption{
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
