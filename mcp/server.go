package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/advancevillage/3rd/logx"
)

const (
	maxScannerBufferSize = 1 << 20 // 1MB
)

var _ Server = (*server)(nil)

type server struct {
	opts      serverOption
	logger    logx.ILogger
	mu        sync.RWMutex
	tools     map[string]Tool
	toolFns   map[string]ToolHandler
	resources map[string]Resource
	resFns    map[string]ResourceHandler
}

// 官方文档
// https://modelcontextprotocol.io/specification/2024-11-05
func NewServer(ctx context.Context, logger logx.ILogger, opt ...ServerOption) Server {
	// 1. 初始化参数
	opts := defaultServerOptions
	for _, o := range opt {
		o.apply(&opts)
	}
	if opts.reader == nil {
		opts.reader = os.Stdin
	}
	if opts.writer == nil {
		opts.writer = os.Stdout
	}

	// 2. 创建服务
	s := &server{
		opts:      opts,
		logger:    logger,
		tools:     make(map[string]Tool),
		toolFns:   make(map[string]ToolHandler),
		resources: make(map[string]Resource),
		resFns:    make(map[string]ResourceHandler),
	}
	logger.Infow(ctx, "mcp: server created", "name", opts.name, "version", opts.version)
	return s
}

func (s *server) RegisterTool(tool Tool, handler ToolHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tools[tool.Name] = tool
	s.toolFns[tool.Name] = handler
}

func (s *server) RegisterResource(resource Resource, handler ResourceHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.resources[resource.URI] = resource
	s.resFns[resource.URI] = handler
}

// Serve 启动 MCP 服务，通过 stdio 传输处理 JSON-RPC 请求
func (s *server) Serve(ctx context.Context) error {
	scanner := bufio.NewScanner(s.opts.reader)
	scanner.Buffer(make([]byte, 0, maxScannerBufferSize), maxScannerBufferSize)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			s.logger.Infow(ctx, "mcp: server stopped")
			return ctx.Err()
		default:
		}

		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		resp := s.handleMessage(ctx, line)
		if resp == nil {
			continue // 通知消息无需响应
		}

		if err := s.writeResponse(resp); err != nil {
			s.logger.Errorw(ctx, "mcp: failed to write response", "err", err)
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		s.logger.Errorw(ctx, "mcp: scanner error", "err", err)
		return err
	}
	return nil
}

func (s *server) handleMessage(ctx context.Context, data []byte) *jsonRPCResponse {
	// 1. 解析请求
	var req jsonRPCRequest
	if err := json.Unmarshal(data, &req); err != nil {
		s.logger.Warnw(ctx, "mcp: parse error", "err", err)
		return s.errorResponse(nil, errCodeParseError, "parse error")
	}

	// 2. 通知消息 (无 ID) 不需要响应
	if req.ID == nil {
		s.logger.Infow(ctx, "mcp: notification received", "method", req.Method)
		return nil
	}

	// 3. 路由请求
	s.logger.Infow(ctx, "mcp: request received", "method", req.Method, "id", req.ID)
	switch req.Method {
	case "initialize":
		return s.handleInitialize(ctx, &req)
	case "ping":
		return s.handlePing(&req)
	case "tools/list":
		return s.handleToolsList(&req)
	case "tools/call":
		return s.handleToolsCall(ctx, &req)
	case "resources/list":
		return s.handleResourcesList(&req)
	case "resources/read":
		return s.handleResourcesRead(ctx, &req)
	default:
		return s.errorResponse(req.ID, errCodeMethodNotFound, fmt.Sprintf("method not found: %s", req.Method))
	}
}

func (s *server) handleInitialize(ctx context.Context, req *jsonRPCRequest) *jsonRPCResponse {
	s.logger.Infow(ctx, "mcp: client initializing")

	// 构建能力集
	capabilities := map[string]any{}
	s.mu.RLock()
	if len(s.tools) > 0 {
		capabilities["tools"] = map[string]any{}
	}
	if len(s.resources) > 0 {
		capabilities["resources"] = map[string]any{}
	}
	s.mu.RUnlock()

	return s.successResponse(req.ID, map[string]any{
		"protocolVersion": protocolVersion,
		"capabilities":    capabilities,
		"serverInfo": map[string]any{
			"name":    s.opts.name,
			"version": s.opts.version,
		},
	})
}

func (s *server) handlePing(req *jsonRPCRequest) *jsonRPCResponse {
	return s.successResponse(req.ID, map[string]any{})
}

func (s *server) handleToolsList(req *jsonRPCRequest) *jsonRPCResponse {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tools := make([]Tool, 0, len(s.tools))
	for _, t := range s.tools {
		tools = append(tools, t)
	}
	return s.successResponse(req.ID, map[string]any{
		"tools": tools,
	})
}

func (s *server) handleToolsCall(ctx context.Context, req *jsonRPCRequest) *jsonRPCResponse {
	// 1. 解析参数
	params, ok := s.extractParams(req)
	if !ok {
		return s.errorResponse(req.ID, errCodeInvalidParams, "invalid params")
	}

	name, _ := params["name"].(string)
	if name == "" {
		return s.errorResponse(req.ID, errCodeInvalidParams, "missing tool name")
	}

	// 2. 查找工具
	s.mu.RLock()
	handler, exists := s.toolFns[name]
	s.mu.RUnlock()

	if !exists {
		return s.errorResponse(req.ID, errCodeInvalidParams, fmt.Sprintf("tool not found: %s", name))
	}

	// 3. 提取工具参数
	args, _ := params["arguments"].(map[string]any)

	// 4. 调用工具
	s.logger.Infow(ctx, "mcp: calling tool", "name", name)
	contents, err := handler(ctx, args)
	if err != nil {
		s.logger.Errorw(ctx, "mcp: tool call failed", "name", name, "err", err)
		return s.successResponse(req.ID, map[string]any{
			"content": []Content{TextContent(fmt.Sprintf("error: %v", err))},
			"isError": true,
		})
	}

	return s.successResponse(req.ID, map[string]any{
		"content": contents,
	})
}

func (s *server) handleResourcesList(req *jsonRPCRequest) *jsonRPCResponse {
	s.mu.RLock()
	defer s.mu.RUnlock()

	resources := make([]Resource, 0, len(s.resources))
	for _, r := range s.resources {
		resources = append(resources, r)
	}
	return s.successResponse(req.ID, map[string]any{
		"resources": resources,
	})
}

func (s *server) handleResourcesRead(ctx context.Context, req *jsonRPCRequest) *jsonRPCResponse {
	// 1. 解析参数
	params, ok := s.extractParams(req)
	if !ok {
		return s.errorResponse(req.ID, errCodeInvalidParams, "invalid params")
	}

	uri, _ := params["uri"].(string)
	if uri == "" {
		return s.errorResponse(req.ID, errCodeInvalidParams, "missing resource uri")
	}

	// 2. 查找资源
	s.mu.RLock()
	handler, exists := s.resFns[uri]
	s.mu.RUnlock()

	if !exists {
		return s.errorResponse(req.ID, errCodeInvalidParams, fmt.Sprintf("resource not found: %s", uri))
	}

	// 3. 读取资源
	s.logger.Infow(ctx, "mcp: reading resource", "uri", uri)
	contents, err := handler(ctx)
	if err != nil {
		s.logger.Errorw(ctx, "mcp: resource read failed", "uri", uri, "err", err)
		return s.errorResponse(req.ID, errCodeInternal, fmt.Sprintf("resource read failed: %v", err))
	}

	return s.successResponse(req.ID, map[string]any{
		"contents": contents,
	})
}

func (s *server) extractParams(req *jsonRPCRequest) (map[string]any, bool) {
	if req.Params == nil {
		return map[string]any{}, true
	}
	params, ok := req.Params.(map[string]any)
	return params, ok
}

func (s *server) successResponse(id any, result any) *jsonRPCResponse {
	return &jsonRPCResponse{
		JSONRPC: jsonRPCVersion,
		ID:      id,
		Result:  result,
	}
}

func (s *server) errorResponse(id any, code int, message string) *jsonRPCResponse {
	return &jsonRPCResponse{
		JSONRPC: jsonRPCVersion,
		ID:      id,
		Error:   &jsonRPCError{Code: code, Message: message},
	}
}

func (s *server) writeResponse(resp *jsonRPCResponse) error {
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	_, err = io.WriteString(s.opts.writer, string(data))
	return err
}
