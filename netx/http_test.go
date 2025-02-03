package netx

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/advancevillage/3rd/logx"
	"github.com/advancevillage/3rd/mathx"
	"github.com/advancevillage/3rd/x"
	"github.com/stretchr/testify/assert"
)

func Test_http(t *testing.T) {
	logger, err := logx.NewLogger("debug")
	assert.Nil(t, err)

	ctx, cancel := context.WithDeadline(context.TODO(), time.Now().Add(time.Second*2))
	go waitQuitSignal(cancel)
	ctx = context.WithValue(ctx, logx.TraceId, mathx.UUID())

	host := "127.0.0.1"
	port := 1995

	opts := []ServerOption{WithServerAddr(host, port)}

	var data = map[string]struct {
		method string
		path   string
		fs     []HttpRegister
	}{
		"case-get": {
			method: http.MethodGet,
			path:   "/resource",
			fs: []HttpRegister{
				func(ctx context.Context, r *http.Request) (HttpResponse, error) {
					logger, err := logx.NewLogger("debug")
					assert.Nil(t, err)
					logger.Infow(ctx, "http-get", "path", r.URL.Path)
					return newHttpResponse(nil, r.Header, http.StatusOK), nil
				},
				func(ctx context.Context, r *http.Request) (HttpResponse, error) {
					logger, err := logx.NewLogger("debug")
					assert.Nil(t, err)

					query := r.URL.Query()
					name := query.Get("name") // 获取 URL 参数

					reply := x.NewBuilder(x.WithKV("name", name)).Build()
					body, err := json.Marshal(reply)
					assert.Nil(t, err)
					time.Sleep(500 * time.Millisecond)
					logger.Infow(ctx, "http-get", "path", r.URL.Path, "name", name)
					return newHttpResponse(body, r.Header, http.StatusOK), nil
				},
			},
		},
	}

	for _, v := range data {
		opts = append(opts, WithHttpService(v.method, v.path, v.fs...))
	}

	s, err := NewHttpServer(ctx, logger, opts...)
	assert.Nil(t, err)
	go s.Start()

	for n, v := range data {
		f := func(t *testing.T) {
			c, err := NewHttpClient(ctx, logger)
			assert.Nil(t, err)
			r, err := c.Get(ctx, fmt.Sprintf("http://t.sunhe.org:%d%s", port, v.path), x.NewBuilder(x.WithKV("name", "pyro")), x.NewBuilder())
			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, r.StatusCode())
			assert.Equal(t, "application/json", r.Header().Get("Content-Type"))
			assert.Equal(t, `{"name":"pyro"}`, string(r.Body()))
		}
		t.Run(n, f)
	}

	<-ctx.Done()
}
