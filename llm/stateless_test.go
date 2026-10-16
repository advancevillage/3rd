package llm_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/advancevillage/3rd/llm"
	"github.com/advancevillage/3rd/logx"
	"github.com/advancevillage/3rd/x"
)

type metadataHandler struct {
	start, end map[string]any
	text       strings.Builder
}

func (h *metadataHandler) OnStart(_ context.Context, opts ...x.Option) {
	h.start = x.NewBuilder(opts...).Build()
}
func (h *metadataHandler) OnEnd(_ context.Context, opts ...x.Option) {
	h.end = x.NewBuilder(opts...).Build()
}
func (h *metadataHandler) OnChunk(_ context.Context, text string) { h.text.WriteString(text) }

// Inspect the actual SDK wire format, including both completion entry points.
func TestStatelessRequests(t *testing.T) {
	for _, mode := range []string{llm.ModeResponse, llm.ModeChat} {
		for _, streaming := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/stream=%v", mode, streaming), func(t *testing.T) {
				requests := make(chan map[string]any, 4)
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					var body map[string]any
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						t.Error(err)
						w.WriteHeader(400)
						return
					}
					requests <- body
					path := "/v1/responses"
					if mode == llm.ModeChat {
						path = "/v1/chat/completions"
					}
					if r.URL.Path != path {
						t.Errorf("path = %s", r.URL.Path)
					}
					if streaming {
						w.Header().Set("Content-Type", "text/event-stream")
						if mode == llm.ModeChat {
							fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"你好\"}}]}\n\ndata: [DONE]\n\n")
						} else {
							fmt.Fprint(w, "data: {\"type\":\"response.created\",\"response\":{\"id\":\"resp_a\"}}\n\n")
							fmt.Fprint(w, "data: {\"type\":\"response.output_text.delta\",\"delta\":\"你好\"}\n\n")
							fmt.Fprint(w, "data: {\"type\":\"response.completed\",\"response\":{\"id\":\"resp_a\",\"usage\":{\"input_tokens\":12,\"output_tokens\":2,\"total_tokens\":14,\"input_tokens_details\":{\"cached_tokens\":8}}}}\n\n")
						}
					} else {
						w.Header().Set("Content-Type", "application/json")
						if mode == llm.ModeChat {
							fmt.Fprint(w, `{"choices":[{"message":{"content":"你好"}}]}`)
						} else {
							fmt.Fprint(w, `{"id":"resp_a","output":[{"type":"message","role":"assistant","content":[{"type":"output_text","text":"你好"}]}]}`)
						}
					}
				}))
				defer server.Close()
				logger, _ := logx.NewLogger("error")
				ctx := context.Background()
				messages := []llm.Message{llm.WithSystemMessage("规则"), llm.WithUserMessage("我叫小明")}
				want := []map[string]any{{"role": "system", "content": "规则"}, {"role": "user", "content": "我叫小明"}}
				// Switch models after receiving a response ID; history remains explicit.
				for round, model := range []string{"model-a", "model-b"} {
					if round == 1 {
						messages = append(messages, llm.WithAssistantMessage("你好"), llm.WithUserMessage("我叫什么？"))
						want = append(want, map[string]any{"role": "assistant", "content": "你好"}, map[string]any{"role": "user", "content": "我叫什么？"})
					}
					opts := []llm.LLMOption{llm.WithBaseUrl(server.URL + "/v1"), llm.WithSecret("test"), llm.WithModel(model), llm.WithMode(mode)}
					if streaming {
						client, err := llm.NewBaseGPT(ctx, logger, opts...)
						if err != nil {
							t.Fatal(err)
						}
						h := &metadataHandler{}
						if err := client.Completion(ctx, h, messages); err != nil {
							t.Fatal(err)
						}
						if h.text.String() != "你好" {
							t.Fatalf("text = %q", h.text.String())
						}
						if mode == llm.ModeResponse {
							if h.start[llm.MetaResponseID] != "resp_a" || h.end[llm.MetaCachedTokens] != int64(8) {
								t.Fatalf("metadata lost: %v %v", h.start, h.end)
							}
							if _, ok := h.start["caching"]; ok {
								t.Fatal("cache flag echoed")
							}
						}
					} else {
						client, err := llm.NewBaseChat(ctx, logger, opts...)
						if err != nil {
							t.Fatal(err)
						}
						reply, err := client.Chat(ctx, messages)
						if err != nil || reply != "你好" {
							t.Fatalf("reply=%q err=%v", reply, err)
						}
					}
					body := <-requests
					for _, key := range []string{"caching", "expire_at", "previous_response_id", "thinking", "tools"} {
						if _, ok := body[key]; ok {
							t.Errorf("unexpected default field %s", key)
						}
					}
					if body["model"] != model {
						t.Errorf("model=%v", body["model"])
					}
					field := "input"
					if mode == llm.ModeChat {
						field = "messages"
					}
					input, ok := body[field].([]any)
					if !ok || len(input) != len(want) {
						t.Fatalf("context lost: %v", body[field])
					}
					for i, expected := range want {
						actual := input[i].(map[string]any)
						for key, value := range expected {
							if !reflect.DeepEqual(actual[key], value) {
								t.Errorf("message %d %s=%v, want %v", i, key, actual[key], value)
							}
						}
					}
				}
			})
		}
	}
}
