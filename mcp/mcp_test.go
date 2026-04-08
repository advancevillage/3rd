package mcp_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/advancevillage/3rd/logx"
	"github.com/advancevillage/3rd/mathx"
	"github.com/advancevillage/3rd/mcp"
	"github.com/stretchr/testify/assert"
)

func newTestServer(ctx context.Context, t *testing.T, input string) (mcp.Server, *bytes.Buffer) {
	t.Helper()
	logger, err := logx.NewLogger("debug")
	assert.Nil(t, err)

	output := &bytes.Buffer{}
	s := mcp.NewServer(ctx, logger,
		mcp.WithServerName("test-server"),
		mcp.WithServerVersion("0.1.0"),
		mcp.WithReader(strings.NewReader(input)),
		mcp.WithWriter(output),
	)
	return s, output
}

func buildRequest(id any, method string, params any) string {
	req := map[string]any{
		"jsonrpc": "2.0",
		"method":  method,
	}
	if id != nil {
		req["id"] = id
	}
	if params != nil {
		req["params"] = params
	}
	data, _ := json.Marshal(req)
	return string(data)
}

func parseResponse(t *testing.T, line string) map[string]any {
	t.Helper()
	var resp map[string]any
	err := json.Unmarshal([]byte(line), &resp)
	assert.Nil(t, err)
	return resp
}

func Test_initialize(t *testing.T) {
	ctx := context.WithValue(context.TODO(), logx.TraceId, mathx.UUID())
	input := buildRequest(1, "initialize", map[string]any{
		"protocolVersion": "2024-11-05",
		"clientInfo":      map[string]any{"name": "test-client", "version": "1.0.0"},
	})

	s, output := newTestServer(ctx, t, input)

	// 注册一个工具确保 capabilities 包含 tools
	s.RegisterTool(mcp.Tool{
		Name:        "echo",
		Description: "echo tool",
		InputSchema: map[string]any{"type": "object"},
	}, func(ctx context.Context, params map[string]any) ([]mcp.Content, error) {
		return []mcp.Content{mcp.TextContent("ok")}, nil
	})

	err := s.Serve(ctx)
	assert.Nil(t, err)

	resp := parseResponse(t, strings.TrimSpace(output.String()))
	assert.Equal(t, "2.0", resp["jsonrpc"])
	assert.Equal(t, float64(1), resp["id"])

	result := resp["result"].(map[string]any)
	assert.Equal(t, "2024-11-05", result["protocolVersion"])
	assert.NotNil(t, result["capabilities"])
	assert.NotNil(t, result["serverInfo"])

	serverInfo := result["serverInfo"].(map[string]any)
	assert.Equal(t, "test-server", serverInfo["name"])
	assert.Equal(t, "0.1.0", serverInfo["version"])

	capabilities := result["capabilities"].(map[string]any)
	assert.NotNil(t, capabilities["tools"])
}

func Test_ping(t *testing.T) {
	ctx := context.WithValue(context.TODO(), logx.TraceId, mathx.UUID())
	input := buildRequest(1, "ping", nil)
	s, output := newTestServer(ctx, t, input)

	err := s.Serve(ctx)
	assert.Nil(t, err)

	resp := parseResponse(t, strings.TrimSpace(output.String()))
	assert.Equal(t, "2.0", resp["jsonrpc"])
	assert.Equal(t, float64(1), resp["id"])
	assert.Nil(t, resp["error"])
}

func Test_tools_list(t *testing.T) {
	ctx := context.WithValue(context.TODO(), logx.TraceId, mathx.UUID())
	input := buildRequest(1, "tools/list", nil)
	s, output := newTestServer(ctx, t, input)

	s.RegisterTool(mcp.Tool{
		Name:        "greet",
		Description: "打招呼工具",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{
					"type":        "string",
					"description": "名字",
				},
			},
			"required": []string{"name"},
		},
	}, func(ctx context.Context, params map[string]any) ([]mcp.Content, error) {
		return nil, nil
	})

	err := s.Serve(ctx)
	assert.Nil(t, err)

	resp := parseResponse(t, strings.TrimSpace(output.String()))
	result := resp["result"].(map[string]any)
	tools := result["tools"].([]any)
	assert.Equal(t, 1, len(tools))

	tool := tools[0].(map[string]any)
	assert.Equal(t, "greet", tool["name"])
	assert.Equal(t, "打招呼工具", tool["description"])
}

func Test_tools_call(t *testing.T) {
	ctx := context.WithValue(context.TODO(), logx.TraceId, mathx.UUID())

	var data = map[string]struct {
		input  string
		tool   mcp.Tool
		fn     mcp.ToolHandler
		expect func(t *testing.T, resp map[string]any)
	}{
		"正常调用": {
			input: buildRequest(1, "tools/call", map[string]any{
				"name":      "add",
				"arguments": map[string]any{"a": 1, "b": 2},
			}),
			tool: mcp.Tool{
				Name:        "add",
				Description: "加法运算",
				InputSchema: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"a": map[string]any{"type": "number"},
						"b": map[string]any{"type": "number"},
					},
				},
			},
			fn: func(ctx context.Context, params map[string]any) ([]mcp.Content, error) {
				a, _ := params["a"].(float64)
				b, _ := params["b"].(float64)
				return []mcp.Content{mcp.TextContent(fmt.Sprintf("%.0f", a+b))}, nil
			},
			expect: func(t *testing.T, resp map[string]any) {
				result := resp["result"].(map[string]any)
				assert.Nil(t, resp["error"])
				contents := result["content"].([]any)
				assert.Equal(t, 1, len(contents))
				c := contents[0].(map[string]any)
				assert.Equal(t, "text", c["type"])
				assert.Equal(t, "3", c["text"])
			},
		},
		"工具返回错误": {
			input: buildRequest(2, "tools/call", map[string]any{
				"name": "fail",
			}),
			tool: mcp.Tool{
				Name:        "fail",
				Description: "失败工具",
				InputSchema: map[string]any{"type": "object"},
			},
			fn: func(ctx context.Context, params map[string]any) ([]mcp.Content, error) {
				return nil, fmt.Errorf("something went wrong")
			},
			expect: func(t *testing.T, resp map[string]any) {
				result := resp["result"].(map[string]any)
				assert.Equal(t, true, result["isError"])
			},
		},
		"工具不存在": {
			input: buildRequest(3, "tools/call", map[string]any{
				"name": "nonexistent",
			}),
			tool: mcp.Tool{
				Name:        "other",
				Description: "其他工具",
				InputSchema: map[string]any{"type": "object"},
			},
			fn: func(ctx context.Context, params map[string]any) ([]mcp.Content, error) {
				return nil, nil
			},
			expect: func(t *testing.T, resp map[string]any) {
				assert.NotNil(t, resp["error"])
			},
		},
	}

	for n, v := range data {
		f := func(t *testing.T) {
			s, output := newTestServer(ctx, t, v.input)
			s.RegisterTool(v.tool, v.fn)
			err := s.Serve(ctx)
			assert.Nil(t, err)
			resp := parseResponse(t, strings.TrimSpace(output.String()))
			v.expect(t, resp)
		}
		t.Run(n, f)
	}
}

func Test_resources_list(t *testing.T) {
	ctx := context.WithValue(context.TODO(), logx.TraceId, mathx.UUID())
	input := buildRequest(1, "resources/list", nil)
	s, output := newTestServer(ctx, t, input)

	s.RegisterResource(mcp.Resource{
		URI:         "file:///config.json",
		Name:        "配置文件",
		Description: "应用配置",
		MimeType:    "application/json",
	}, func(ctx context.Context) ([]mcp.Content, error) {
		return nil, nil
	})

	err := s.Serve(ctx)
	assert.Nil(t, err)

	resp := parseResponse(t, strings.TrimSpace(output.String()))
	result := resp["result"].(map[string]any)
	resources := result["resources"].([]any)
	assert.Equal(t, 1, len(resources))

	res := resources[0].(map[string]any)
	assert.Equal(t, "file:///config.json", res["uri"])
	assert.Equal(t, "配置文件", res["name"])
}

func Test_resources_read(t *testing.T) {
	ctx := context.WithValue(context.TODO(), logx.TraceId, mathx.UUID())
	input := buildRequest(1, "resources/read", map[string]any{
		"uri": "file:///config.json",
	})
	s, output := newTestServer(ctx, t, input)

	s.RegisterResource(mcp.Resource{
		URI:      "file:///config.json",
		Name:     "配置文件",
		MimeType: "application/json",
	}, func(ctx context.Context) ([]mcp.Content, error) {
		return []mcp.Content{mcp.TextContent(`{"key":"value"}`)}, nil
	})

	err := s.Serve(ctx)
	assert.Nil(t, err)

	resp := parseResponse(t, strings.TrimSpace(output.String()))
	result := resp["result"].(map[string]any)
	contents := result["contents"].([]any)
	assert.Equal(t, 1, len(contents))

	c := contents[0].(map[string]any)
	assert.Equal(t, "text", c["type"])
	assert.Equal(t, `{"key":"value"}`, c["text"])
}

func Test_notification(t *testing.T) {
	ctx := context.WithValue(context.TODO(), logx.TraceId, mathx.UUID())
	// 通知消息没有 id，不应有响应
	input := buildRequest(nil, "notifications/initialized", nil)
	s, output := newTestServer(ctx, t, input)

	err := s.Serve(ctx)
	assert.Nil(t, err)
	assert.Equal(t, "", output.String())
}

func Test_method_not_found(t *testing.T) {
	ctx := context.WithValue(context.TODO(), logx.TraceId, mathx.UUID())
	input := buildRequest(1, "unknown/method", nil)
	s, output := newTestServer(ctx, t, input)

	err := s.Serve(ctx)
	assert.Nil(t, err)

	resp := parseResponse(t, strings.TrimSpace(output.String()))
	assert.NotNil(t, resp["error"])
	errObj := resp["error"].(map[string]any)
	assert.Equal(t, float64(-32601), errObj["code"])
}

func Test_multi_request(t *testing.T) {
	ctx := context.WithValue(context.TODO(), logx.TraceId, mathx.UUID())
	lines := []string{
		buildRequest(1, "initialize", map[string]any{
			"protocolVersion": "2024-11-05",
			"clientInfo":      map[string]any{"name": "test", "version": "1.0"},
		}),
		buildRequest(nil, "notifications/initialized", nil),
		buildRequest(2, "tools/list", nil),
		buildRequest(3, "ping", nil),
	}
	input := strings.Join(lines, "\n")
	s, output := newTestServer(ctx, t, input)

	err := s.Serve(ctx)
	assert.Nil(t, err)

	// 应有 3 个响应 (notification 无响应)
	responses := strings.Split(strings.TrimSpace(output.String()), "\n")
	assert.Equal(t, 3, len(responses))
}
