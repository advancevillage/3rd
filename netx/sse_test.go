package netx

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/advancevillage/3rd/logx"
	"github.com/advancevillage/3rd/mathx"
	"github.com/advancevillage/3rd/x"
	"github.com/stretchr/testify/assert"
)

func Test_sse(t *testing.T) {
	logger, err := logx.NewLogger("debug")
	assert.Nil(t, err)

	ctx, cancel := context.WithDeadline(context.TODO(), time.Now().Add(time.Second*2))
	go waitQuitSignal(cancel)
	ctx = context.WithValue(ctx, logx.TraceId, mathx.UUID())

	host := "127.0.0.1"
	port := 1994

	opts := []ServerOption{WithServerAddr(host, port)}

	var data = map[string]struct {
		method string
		path   string
		fs     []HttpRegister
	}{
		"case-sse": {
			method: http.MethodGet,
			path:   "/sse",
			fs: []HttpRegister{
				func(ctx context.Context, r *http.Request) (HttpResponse, error) {
					logger.Infow(ctx, "http-get-01", "path", r.URL.Path)
					return NewContextResponse(
						x.WithKV("account", "niuniu"),
					), nil
				},
			},
		},
	}

	for _, v := range data {
		opts = append(opts, WithHttpService(v.method, v.path, append(v.fs, NewSSESrv(ctx, logger))...))
	}

	s, err := NewHttpServer(ctx, logger, opts...)
	assert.Nil(t, err)
	go s.Start()
	time.Sleep(time.Second * 2)

	for n, v := range data {
		f := func(t *testing.T) {
			c, err := NewHttpClient(ctx, logger, WithClientTimeout(3600))
			assert.Nil(t, err)
			r, err := c.Get(ctx, fmt.Sprintf("http://%s:%d%s", host, port, v.path), x.NewBuilder(x.WithKV("name", "pyro")), x.NewBuilder())
			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, r.StatusCode())
			assert.Equal(t, "text/event-stream", r.Header().Get("Content-Type"))
		}
		t.Run(n, f)
	}

	<-ctx.Done()
}
