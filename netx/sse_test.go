package netx

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/advancevillage/3rd/logx"
	"github.com/advancevillage/3rd/mathx"
	"github.com/stretchr/testify/assert"
)

type testSSEHandler struct {
	t *testing.T
}

func (t *testSSEHandler) OnChunk(ctx context.Context, r *http.Request) <-chan SSEvent {
	query := r.URL.Query()
	name := query.Get("name")
	assert.Equal(t.t, "pyro", name)

	ch := make(chan SSEvent)
	go func() {
		c := 0
		tr := time.NewTicker(time.Second)
		defer tr.Stop()

		for {
			select {
			case <-ctx.Done():
				close(ch)
				return
			case <-tr.C:
				c += 1
				ch <- NewSSEvent(mathx.RandStr(10), time.Now().Format("2006-01-02 15:04:05"))
			}
			if c >= 10 {
				close(ch)
				return
			}
		}
	}()
	return ch
}

type testSSEParser struct {
	id    string
	event string
	data  strings.Builder
}

func (p *testSSEParser) ParseLine(raw string) (*sseEvent, bool) {
	line := strings.TrimSpace(raw)

	switch {
	// 空行：一个事件结束
	case len(line) == 0 && p.data.Len() == 0:
		return nil, false

	// 空行：一个事件结束
	case len(line) == 0 && p.data.Len() > 0:
		evt := &sseEvent{
			id:    p.id,
			event: p.event,
			data:  p.data.String(),
		}
		// reset (每个 SSE event 都要重置)
		p.id = ""
		p.event = ""
		p.data.Reset()
		return evt, true

	case strings.HasPrefix(line, "event:"):
		p.event = strings.TrimSpace(line[len("event:"):])

	case strings.HasPrefix(line, "id:"):
		p.id = strings.TrimSpace(line[len("id:"):])

	case strings.HasPrefix(line, "data:"):
		data := strings.TrimSpace(line[len("data:"):])
		if p.data.Len() > 0 {
			p.data.WriteByte('\n')
		}
		p.data.WriteString(data)
	}
	return nil, false
}

func Test_sse(t *testing.T) {
	logger, err := logx.NewLogger("debug")
	assert.Nil(t, err)

	ctx, cancel := context.WithDeadline(context.TODO(), time.Now().Add(time.Second*12))
	go waitQuitSignal(cancel)
	ctx = context.WithValue(ctx, logx.TraceId, mathx.UUID())

	host := "127.0.0.1"
	port := 1994

	opts := []ServerOption{WithServerAddr(host, port)}
	handler := &testSSEHandler{t: t}

	var data = map[string]struct {
		method string
		path   string
		fs     []HttpRegister
	}{
		"case-sse": {
			method: http.MethodGet,
			path:   "/sse",
			fs:     []HttpRegister{},
		},
	}

	for _, v := range data {
		opts = append(opts, WithHttpService(v.method, v.path, append(v.fs, NewSSESrv(ctx, logger, WithSSEventHandler(handler.OnChunk)))...))
	}

	s, err := NewHttpServer(ctx, logger, opts...)
	assert.Nil(t, err)
	go s.Start()
	time.Sleep(time.Second * 2)

	for n, v := range data {
		f := func(t *testing.T) {
			req, err := http.NewRequest("GET", fmt.Sprintf("http://%s:%d%s?name=pyro", host, port, v.path), nil)
			assert.Nil(t, err)
			req.Header.Set("Accept", "text/event-stream")
			req.WithContext(ctx)

			resp, err := http.DefaultClient.Do(req)
			assert.Nil(t, err)
			defer resp.Body.Close()

			reader := bufio.NewReader(resp.Body)
			parser := &testSSEParser{}
			c := 0
			for {
				line, err := reader.ReadString('\n')
				if errors.Is(err, io.EOF) {
					break
				} else {
					assert.Nil(t, err)
				}
				if evt, ok := parser.ParseLine(line); ok {
					assert.Equal(t, evt.id, fmt.Sprint(c))
					c += 1
					t.Logf("SSE event: id=%s event=%s data=%s", evt.id, evt.event, evt.data)
				}
			}
		}
		t.Run(n, f)
	}

	<-ctx.Done()
}
