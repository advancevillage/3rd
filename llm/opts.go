package llm

const (
	roleUser   = "user"
	roleSystem = "system"
	roleAssist = "assistant"
)

// Completion API 方式
const (
	ModeResponse = "response" // Response API（默认）
	ModeChat     = "chat"     // Chat Completion API
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

// commonOption 抽取所有 Option 共有的字段，作为基类嵌入各具体 Option，
// set* 方法是这些字段唯一的写入实现。
// baseUrl 统一原 host/dsn(服务地址)，timeout 统一原 timeout/ttl(时长，单位秒)。
type commonOption struct {
	sk      string
	model   string
	baseUrl string
	timeout int
}

func (o *commonOption) setSecret(sk string)   { o.sk = sk }
func (o *commonOption) setModel(model string) { o.model = model }
func (o *commonOption) setBaseUrl(url string) { o.baseUrl = url }
func (o *commonOption) setTimeout(t int)      { o.timeout = t }

// 通用 setter 接口: 由具体 Option 嵌入的 commonOption 统一实现，供泛型 with* 复用。
type (
	secretSetter  interface{ setSecret(string) }
	modelSetter   interface{ setModel(string) }
	baseUrlSetter interface{ setBaseUrl(string) }
	timeoutSetter interface{ setTimeout(int) }
)

// 泛型核心: T 为具体 Option 值类型，PT 经约束类型推导为 *T。
// 调用方只需指定 T(如 withModel[llmOption])，PT 由编译器自动推导。
func withSecret[T any, PT interface {
	*T
	secretSetter
}](sk string) Option[T] {
	return newFuncOption(func(o *T) { PT(o).setSecret(sk) })
}

func withModel[T any, PT interface {
	*T
	modelSetter
}](model string) Option[T] {
	return newFuncOption(func(o *T) { PT(o).setModel(model) })
}

func withBaseUrl[T any, PT interface {
	*T
	baseUrlSetter
}](url string) Option[T] {
	return newFuncOption(func(o *T) { PT(o).setBaseUrl(url) })
}

func withTimeout[T any, PT interface {
	*T
	timeoutSetter
}](t int) Option[T] {
	return newFuncOption(func(o *T) { PT(o).setTimeout(t) })
}

type LLMOption = Option[llmOption]

type llmOption struct {
	commonOption
	// mode Completion API 方式: response(默认) / chat
	mode string
}

var defaultLLMOptions = llmOption{
	commonOption: commonOption{
		model:   "doubao-seed-2-0-lite-260428",
		baseUrl: "https://ark.cn-beijing.volces.com/api/v3",
	},
	mode: ModeResponse,
}

func WithBaseUrl(url string) LLMOption { return withBaseUrl[llmOption](url) }

func WithModel(model string) LLMOption { return withModel[llmOption](model) }

func WithSecret(sk string) LLMOption { return withSecret[llmOption](sk) }

func WithMode(mode string) LLMOption {
	return newFuncOption(func(o *llmOption) { o.mode = mode })
}

// CompletionOption 每次 Completion 调用级别的透传选项, 仅 Response API 路径生效。
type CompletionOption = Option[completionOption]

type completionOption struct {
	// webSearch 开启联网检索工具: tools=[{"type":"web_search"}]
	webSearch bool
	// thinking 开启思考模式: thinking={"type":"enabled"}
	thinking bool
	// caching 开启上下文缓存: caching={"type":"enabled"}, 命中重置
	caching bool
	// expireAt 缓存过期时长(秒), 经 expire_at=now+expireAt 显式控制, 默认 2 小时
	expireAt int64
	// prevRespId 复用上一轮缓存: previous_response_id
	prevRespId string
}

// defaultCacheTTLSeconds 上下文缓存默认过期时长: 2 小时。
const defaultCacheTTLSeconds int64 = 7200

// WithWebSearch 开启联网检索。
func WithWebSearch() CompletionOption {
	return newFuncOption(func(o *completionOption) { o.webSearch = true })
}

// WithThinking 开启思考模式。
func WithThinking() CompletionOption {
	return newFuncOption(func(o *completionOption) { o.thinking = true })
}

// WithCache 开启上下文缓存, 本轮响应会被缓存, 过期时长默认 2 小时。
func WithCache() CompletionOption {
	return newFuncOption(func(o *completionOption) { o.caching = true; o.expireAt = defaultCacheTTLSeconds })
}

// WithCacheTTL 开启上下文缓存并指定过期时长(秒), 上限 259200(72 小时)。
func WithCacheTTL(seconds int64) CompletionOption {
	return newFuncOption(func(o *completionOption) { o.caching = true; o.expireAt = seconds })
}

// WithPreviousResponse 携带上一轮的 response id 命中缓存。
func WithPreviousResponse(id string) CompletionOption {
	return newFuncOption(func(o *completionOption) { o.prevRespId = id })
}

type LLMStreamOption = Option[llmStreamOption]

type llmStreamOption struct {
	commonOption
	region  string
	handler StreamHandler
}

var defaultLLMStreamOptions = llmStreamOption{
	commonOption: commonOption{timeout: 600},
	region:       "ap-guangzhou",
	handler:      &emptyStreamHandler{},
}

func WithStreamSecret(sk string) LLMStreamOption { return withSecret[llmStreamOption](sk) }

func WithStreamModel(model string) LLMStreamOption { return withModel[llmStreamOption](model) }

func WithStreamTimeout(timeout int) LLMStreamOption { return withTimeout[llmStreamOption](timeout) }

func WithStreamHandler(handler StreamHandler) LLMStreamOption {
	return newFuncOption(func(o *llmStreamOption) {
		o.handler = handler
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
