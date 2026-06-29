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
}

var defaultLLMOptions = llmOption{
	commonOption: commonOption{model: "gpt-5-mini", baseUrl: "https://api.openai.com/v1"},
}

func WithBaseUrl(url string) LLMOption { return withBaseUrl[llmOption](url) }

func WithModel(model string) LLMOption { return withModel[llmOption](model) }

func WithSecret(sk string) LLMOption { return withSecret[llmOption](sk) }

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

// 上下文缓存(Context)
type LLMContextOption = Option[llmContextOption]

type llmContextOption struct {
	commonOption
	// mode 缓存模式: session(每轮动态更新, 多轮对话) / common_prefix(前缀缓存, 静态不更新)
	mode string
	// ttl 缓存存活时间(秒), 未命中倒计时, 命中则重置
	ttl int
	// maxWindowTokens session 模式 rolling_tokens 截断上限, 缓存达此阈值触发截断
	maxWindowTokens int
	// rollingWindowTokens session 模式 rolling_tokens 触发截断时滚动删除的陈旧历史 token 数
	rollingWindowTokens int
}

var defaultLLMContextOptions = llmContextOption{
	commonOption:        commonOption{model: "doubao-seed-2-0-mini-260428", baseUrl: "https://ark.cn-beijing.volces.com/api/v3"},
	mode:                "session",
	ttl:                 3600,
	maxWindowTokens:     8192,
	rollingWindowTokens: 4096,
}

func WithContextModel(model string) LLMContextOption { return withModel[llmContextOption](model) }

func WithContextSecret(sk string) LLMContextOption { return withSecret[llmContextOption](sk) }

func WithContextBaseUrl(url string) LLMContextOption { return withBaseUrl[llmContextOption](url) }

func WithContextTimeout(timeout int) LLMContextOption { return withTimeout[llmContextOption](timeout) }

func WithContextMode(mode string) LLMContextOption {
	return newFuncOption(func(o *llmContextOption) { o.mode = mode })
}

func WithContextTTL(ttl int) LLMContextOption {
	return newFuncOption(func(o *llmContextOption) { o.ttl = ttl })
}

func WithContextMaxWindowTokens(tokens int) LLMContextOption {
	return newFuncOption(func(o *llmContextOption) { o.maxWindowTokens = tokens })
}

func WithContextRollingWindowTokens(tokens int) LLMContextOption {
	return newFuncOption(func(o *llmContextOption) { o.rollingWindowTokens = tokens })
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
