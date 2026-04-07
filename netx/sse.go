package netx

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/advancevillage/3rd/logx"
	"github.com/gin-gonic/gin"
)

const (
	ErrorSSEventType   = "error"
	HeaderSSEventType  = "header"
	MessageSSEventType = "message"
)

type SSEvent interface {
	Data() string
	Event() string
}

var _ SSEvent = (*sseEvent)(nil)

type sseEvent struct {
	id    string
	data  string
	event string
}

func NewSSEvent(event string, data string) SSEvent {
	return &sseEvent{
		data:  data,
		event: event,
	}
}

func NewHeaderSSEvent(data string) SSEvent {
	return NewSSEvent(HeaderSSEventType, data)
}

func NewMessageSSEvent(data string) SSEvent {
	return NewSSEvent(MessageSSEventType, data)
}

func NewErrorSSEvent(data string) SSEvent {
	return NewSSEvent(ErrorSSEventType, data)
}

func (c *sseEvent) Data() string {
	return c.data
}

func (c *sseEvent) Event() string {
	return c.event
}

type SSEventHandler func(ctx context.Context, r *http.Request) <-chan SSEvent

type SSEventOption = Option[sseOptions]

func WithSSEventHandler(handler SSEventHandler) SSEventOption {
	return newFuncOption(func(o *sseOptions) {
		o.handler = handler
	})
}

func WithSSEventProxyMode() SSEventOption {
	return newFuncOption(func(o *sseOptions) {
		o.proxy = true
	})
}

type sseOptions struct {
	handler SSEventHandler
	proxy   bool
}

var defaultSSEOptions = sseOptions{
	handler: emptySSEventHandler,
	proxy:   false,
}

var emptySSEventHandler = func(ctx context.Context, r *http.Request) <-chan SSEvent {
	ch := make(chan SSEvent, 1)
	go func() {
		close(ch)
	}()
	return ch
}

type sseSrv struct {
	opts   sseOptions
	logger logx.ILogger
}

func NewSSESrv(ctx context.Context, logger logx.ILogger, opt ...SSEventOption) HttpRegister {
	// 1. 设置配置
	opts := defaultSSEOptions
	for _, o := range opt {
		o.apply(&opts)
	}
	// 2. 创建服务
	s := &sseSrv{
		opts:   opts,
		logger: logger,
	}
	logger.Infow(ctx, "sse: server created", "handler", opts.handler != nil)
	// 3. 返回Http注册路由
	if opts.proxy {
		return s.proxy // 中转透传模式
	} else {
		return s.stream // 标准SSE模式
	}
}

func (s *sseSrv) stream(ctx context.Context, r *http.Request) (HttpResponse, error) {
	// 1. 定义数据
	h := http.Header{}
	h.Add("Content-Type", "text/event-stream")
	h.Add("Cache-Control", "no-cache")
	h.Add("Connection", "keep-alive")
	h.Add("Transfer-Encoding", "chunked")

	// 1. 透传数据，上游处理
	writer, ok := ctx.Value(ctxKeyResponseWriter{}).(gin.ResponseWriter)
	if !ok {
		return newHttpResponse([]byte("sse: no response writer"), h, http.StatusInternalServerError), nil
	}
	for k, v := range h {
		writer.Header().Add(k, v[0])
	}

	replyFunc := func(evtId int, event string, data string) HttpResponse {
		return newHttpResponse(s.pack(evtId, event, data), h, http.StatusOK)
	}

	// 2. 读取请求体
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return replyFunc(0, "error", fmt.Sprintf("err=%v", err)), nil
	}
	r.Body.Close()

	// 3. 起始消息(会将r.Body数据清空)
	n, err := writer.Write(s.pack(0, "open", "welcome"))
	if err != nil {
		return replyFunc(0, "error", fmt.Sprintf("n=%d err=%v", n, err)), nil
	}
	writer.Flush()

	// 4. 重建body
	r.Body = io.NopCloser(bytes.NewBuffer(body))

	// 5. 处理数据请求
	var (
		id     = 1
		events = s.opts.handler(ctx, r)
	)

	// 6. 循环发送数据
	for {
		select {
		case <-ctx.Done():
			s.logger.Infow(ctx, "sse: client closed", "id", id)
			goto exitLoop

		case evt, ok := <-events:
			if ok {
				n, err = writer.Write(s.pack(id, evt.Event(), evt.Data()))
				if err != nil {
					return replyFunc(id, "error", fmt.Sprintf("n=%d err=%v", n, err)), nil
				}
			} else {
				goto exitLoop
			}
		}
		writer.Flush()
		id += 1
	}

	// 7. 关闭连接
exitLoop:
	return replyFunc(id, "close", "bye"), nil
}

func (s *sseSrv) pack(id int, event string, data string) []byte {
	p := fmt.Sprintf("id:%d\n", id)
	p += fmt.Sprintf("event:%s\n", event)
	p += fmt.Sprintf("data:%s\n\n", data)
	return []byte(p)
}

type ctxKeySSEStream struct{}

func WithSSEStream(ctx context.Context, stream bool) context.Context {
	return context.WithValue(ctx, ctxKeySSEStream{}, stream)
}

func (s *sseSrv) proxy(ctx context.Context, r *http.Request) (HttpResponse, error) {
	stream, ok := ctx.Value(ctxKeySSEStream{}).(bool)
	if stream && ok {
		return s.proxyStream(ctx, r)
	}
	return s.proxyDirect(ctx, r)
}

func (s *sseSrv) proxyStream(ctx context.Context, r *http.Request) (HttpResponse, error) {
	writer, ok := ctx.Value(ctxKeyResponseWriter{}).(gin.ResponseWriter)
	if !ok {
		return newHttpResponse([]byte("sse: no response writer"), http.Header{}, http.StatusInternalServerError), nil
	}

	defaultHeaders := http.Header{}
	defaultHeaders.Set("Content-Type", "text/event-stream")
	defaultHeaders.Set("Cache-Control", "no-cache")
	defaultHeaders.Set("Connection", "keep-alive")
	defaultHeaders.Set("X-Accel-Buffering", "no")

	events := s.opts.handler(ctx, r)
	firstChunk := true

	for {
		select {
		case <-ctx.Done():
			s.logger.Infow(ctx, "sse: client closed")
			goto exitLoop

		case evt, ok := <-events:
			if !ok {
				goto exitLoop
			}

			// header 事件只更新待写 headers，不立即发送
			if evt.Event() == HeaderSSEventType {
				h := s.parseHeaderEvent(ctx, evt.Data())
				if len(h) > 0 {
					defaultHeaders = h
				}
				continue
			}

			// 首包：确保 response headers 在任何数据写入前发出
			if firstChunk {
				firstChunk = false
				for k, v := range defaultHeaders {
					writer.Header().Set(k, strings.Join(v, ","))
				}
				if evt.Event() == ErrorSSEventType {
					writer.WriteHeader(http.StatusInternalServerError)
				} else {
					writer.WriteHeader(http.StatusOK)
				}
				writer.Flush()
			}

			if evt.Event() == ErrorSSEventType {
				goto exitLoop
			}

			n, err := writer.WriteString(evt.Data())
			if err != nil {
				s.logger.Errorw(ctx, "sse: write error", "n", n, "err", err)
				goto exitLoop
			}
			writer.Flush()
		}
	}

exitLoop:
	return newHttpResponse([]byte{}, http.Header{}, http.StatusOK), nil
}

func (s *sseSrv) proxyDirect(ctx context.Context, r *http.Request) (HttpResponse, error) {
	var buf bytes.Buffer
	events := s.opts.handler(ctx, r)
	h := http.Header{}
	statusCode := http.StatusOK

	for {
		select {
		case <-ctx.Done():
			s.logger.Infow(ctx, "sse: client closed")
			goto exitLoop

		case evt, ok := <-events:
			if !ok {
				goto exitLoop
			}

			switch evt.Event() {
			case HeaderSSEventType:
				h = s.parseHeaderEvent(ctx, evt.Data())

			case ErrorSSEventType:
				statusCode = http.StatusInternalServerError
				goto exitLoop

			default:
				buf.WriteString(evt.Data())
			}
		}
	}

exitLoop:
	return newHttpResponse(buf.Bytes(), h, statusCode), nil
}

func (s *sseSrv) parseHeaderEvent(ctx context.Context, raw string) http.Header {
	q, err := url.ParseQuery(raw)
	if err != nil {
		s.logger.Errorw(ctx, "sse: parse header event error", "raw", raw, "err", err)
		return http.Header{}
	}
	h := http.Header{}
	for k, v := range q {
		h.Set(k, strings.Join(v, ","))
	}
	return h
}
