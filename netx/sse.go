package netx

import (
	"context"
	"fmt"
	"net/http"

	"github.com/advancevillage/3rd/logx"
	"github.com/gin-gonic/gin"
)

type SSEvent interface {
	Data() string
	Event() string
}

type SSEventHandler func(ctx context.Context, r *http.Request) <-chan SSEvent

type SSEventOption = Option[sseOptions]

func WithSSEventHandler(handler SSEventHandler) SSEventOption {
	return newFuncOption(func(o *sseOptions) {
		o.handler = handler
	})

}

type sseOptions struct {
	handler SSEventHandler
}

var defaultSSEOptions = sseOptions{
	handler: emptySSEventHandler,
}

var emptySSEventHandler = func(ctx context.Context, r *http.Request) <-chan SSEvent {
	events := make(chan SSEvent)
	close(events)
	return events
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
	// 3. 返回Http注册路由
	return s.stream
}

func (s *sseSrv) stream(ctx context.Context, r *http.Request) (HttpResponse, error) {
	// 1. 定义数据
	h := http.Header{}
	h.Add("Content-Type", "text/event-stream")
	h.Add("Cache-Control", "no-cache")
	h.Add("Connection", "keep-alive")

	// 1. 透传数据，上游处理
	writer, ok := ctx.Value(ctxKeyResponseWriter{}).(gin.ResponseWriter)
	if !ok {
		return newHttpResponse([]byte("sse: no response writer"), h, http.StatusInternalServerError), nil
	}
	writer.Header().Add("Content-Type", "text/event-stream")
	writer.Header().Add("Cache-Control", "no-cache")
	writer.Header().Add("Connection", "keep-alive")

	// 2. 起始消息
	writer.Write(s.pack(0, "open", "welcome"))
	writer.Flush()

	// 3. 创建通道用于接收关闭通知
	var (
		id           = 1
		events       = s.opts.handler(ctx, r)
		clientClosed = writer.CloseNotify()
	)

	// 4. 循环发送数
	for {
		select {
		case <-clientClosed:
			s.logger.Infow(ctx, "sse: client closed", "id", id)
			goto exitLoop

		case evt, ok := <-events:
			if ok {
				writer.Write(s.pack(id, evt.Event(), evt.Data()))
			} else {
				goto exitLoop
			}
		}
		writer.Flush()
		id += 1
	}

	// 5. 关闭连接
exitLoop:
	return newHttpResponse(s.pack(id, "close", "bye"), h, http.StatusOK), nil
}

func (s *sseSrv) pack(id int, event string, data string) []byte {
	p := fmt.Sprintf("id:%d\n", id)
	p += fmt.Sprintf("event:%s\n", event)
	p += fmt.Sprintf("data:%s\n\n", data)
	return []byte(p)
}
