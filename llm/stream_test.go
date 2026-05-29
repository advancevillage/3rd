package llm

import (
	"context"
	"testing"
)

type collectStreamHandler struct {
	chunks []string
}

func (h *collectStreamHandler) OnStart(context.Context) {}
func (h *collectStreamHandler) OnChunk(_ context.Context, chunk string) {
	h.chunks = append(h.chunks, chunk)
}
func (h *collectStreamHandler) OnEnd(context.Context) {}

func TestBufferStreamHandlerOnChunkEmitToLastSeparator(t *testing.T) {
	handler := &collectStreamHandler{}
	h := &bufferStreamHandler{handler: handler}

	h.OnChunk(context.Background(), "你好！很好！继续")

	if len(handler.chunks) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(handler.chunks))
	}
	if handler.chunks[0] != "你好！很好！" {
		t.Fatalf("expected chunk %q, got %q", "你好！很好！", handler.chunks[0])
	}
	if h.buf != "继续" {
		t.Fatalf("expected remaining buf %q, got %q", "继续", h.buf)
	}
}

func TestBufferStreamHandlerOnChunkEscapeNewline(t *testing.T) {
	handler := &collectStreamHandler{}
	h := &bufferStreamHandler{handler: handler}

	h.OnChunk(context.Background(), "结束。\n下一句")

	if len(handler.chunks) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(handler.chunks))
	}
	if handler.chunks[0] != "结束。\\n" {
		t.Fatalf("expected chunk %q, got %q", "结束。\\n", handler.chunks[0])
	}
	if h.buf != "下一句" {
		t.Fatalf("expected remaining buf %q, got %q", "下一句", h.buf)
	}
}
