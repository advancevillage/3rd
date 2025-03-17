package llm

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
		o.model = "o3-mini"
	})
}

func With4OModel() LLMOption {
	return newFuncLLMOption(func(o *llmOption) {
		o.model = "gpt-4o"
	})
}

type llmOption struct {
	sk    string
	model string
}

var defaultLLMOptions = llmOption{
	model: "gpt-4o-mini",
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
