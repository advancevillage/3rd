package llm

const (
	roleUser   = "user"
	roleSystem = "system"
	roleAssist = "assistant"
)

type Option[T any] interface {
	apply(*T)
}

type funcOption[T any] struct {
	f func(*T)
}

func (o funcOption[T]) apply(do *T) {
	o.f(do)
}

func newFuncOption[T any](f func(*T)) *funcOption[T] {
	return &funcOption[T]{f: f}
}

type LLMOption = Option[llmOption]

type llmOption struct {
	sk    string
	proxy string
	model string
}

var defaultLLMOptions = llmOption{
	model: "4o-mini",
}

func WithModel(model string) LLMOption {
	return newFuncOption(func(o *llmOption) {
		o.model = model
	})
}

func WithChatGPTSecret(sk string) LLMOption {
	return newFuncOption(func(o *llmOption) {
		o.sk = sk
	})
}

func WithChatGPTProxy(proxy string) LLMOption {
	return newFuncOption(func(o *llmOption) {
		o.proxy = proxy
	})
}

type LLMStreamOption = Option[llmStreamOption]

type llmStreamOption struct {
	ak      string
	sk      string
	model   string
	region  string
	handler StreamHandler
	timeout int
}

var defaultLLMStreamOptions = llmStreamOption{
	region:  "ap-guangzhou",
	handler: &emptyStreamHandler{},
	timeout: 600,
}

func WithStreamSecret(ak string, sk string) LLMStreamOption {
	return newFuncOption(func(o *llmStreamOption) {
		o.ak = ak
		o.sk = sk
	})
}

func WithStreamModel(model string) LLMStreamOption {
	return newFuncOption(func(o *llmStreamOption) {
		o.model = model
	})
}

func WithStreamHandler(handler StreamHandler) LLMStreamOption {
	return newFuncOption(func(o *llmStreamOption) {
		o.handler = handler
	})
}

func WithStreamTimeout(timeout int) LLMStreamOption {
	return newFuncOption(func(o *llmStreamOption) {
		o.timeout = timeout
	})
}

func WithStreamRegion(region string) LLMStreamOption {
	return newFuncOption(func(o *llmStreamOption) {
		o.region = region
	})
}

// 消息优先级: system > user > assistant(历史)
type Message = Option[message]

type message struct {
	role    string
	content string
}

// user — 用户的真实需求（普通优先级）
func WithUserMessage(content string) Message {
	return newFuncOption(func(m *message) {
		m.role = roleUser
		m.content = content
	})
}

// system — 角色设定、全局规则（最高级别）
func WithSystemMessage(content string) Message {
	return newFuncOption(func(m *message) {
		m.role = roleSystem
		m.content = content
	})
}

// assistant — 模型输出
func WithAssistantMessage(content string) Message {
	return newFuncOption(func(m *message) {
		m.role = roleAssist
		m.content = content
	})
}
