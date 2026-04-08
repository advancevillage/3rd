package mcp

import "context"

// JSON-RPC 2.0 协议版本
const jsonRPCVersion = "2.0"

// MCP 协议版本
const protocolVersion = "2024-11-05"

// JSON-RPC 2.0 请求
type jsonRPCRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id,omitempty"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

// JSON-RPC 2.0 响应
type jsonRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      any           `json:"id,omitempty"`
	Result  any           `json:"result,omitempty"`
	Error   *jsonRPCError `json:"error,omitempty"`
}

// JSON-RPC 2.0 错误
type jsonRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// JSON-RPC 2.0 标准错误码
const (
	errCodeParseError     = -32700
	errCodeInvalidRequest = -32600
	errCodeMethodNotFound = -32601
	errCodeInvalidParams  = -32602
	errCodeInternal       = -32603
)

// MCP 工具定义
type Tool struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	InputSchema map[string]any `json:"inputSchema"`
}

// MCP 资源定义
type Resource struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	MimeType    string `json:"mimeType,omitempty"`
}

// MCP 内容类型
type Content struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	MimeType string `json:"mimeType,omitempty"`
	Data     string `json:"data,omitempty"`
}

// 创建文本内容
func TextContent(text string) Content {
	return Content{Type: "text", Text: text}
}

// 工具处理函数
type ToolHandler func(ctx context.Context, params map[string]any) ([]Content, error)

// 资源处理函数
type ResourceHandler func(ctx context.Context) ([]Content, error)

// MCP 服务端接口
type Server interface {
	// 注册工具
	RegisterTool(tool Tool, handler ToolHandler)
	// 注册资源
	RegisterResource(resource Resource, handler ResourceHandler)
	// 启动服务 (stdio 传输)
	Serve(ctx context.Context) error
}
