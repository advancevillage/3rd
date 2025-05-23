package netx

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
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
					logger.Infow(ctx, "http-get-01", "path", r.URL.Path)
					return NewContextResponse(
						x.WithKV("account", "niuniu"),
						x.WithKV("niu-niu", "3 year old"),
						x.WithKV("niu_niu", "3-year_old"),
						x.WithKV("niu_niu+34", "3_year+old"),
						x.WithKV("accountId", "acc-1891362318018039808"),
						x.WithKV("ABC", "acc-1891362318018039808"),
					), nil
				},
				func(ctx context.Context, r *http.Request) (HttpResponse, error) {
					logger.Infow(ctx, "http-get-02", "path", r.URL.Path)
					assert.Equal(t, "niuniu", ctx.Value("account"))
					assert.Equal(t, "3 year old", ctx.Value("niu-niu"))
					assert.Equal(t, "3-year_old", ctx.Value("niu_niu"))
					assert.Equal(t, "3_year+old", ctx.Value("niu_niu+34"))
					assert.Equal(t, "acc-1891362318018039808", ctx.Value("accountId"))
					assert.Equal(t, "acc-1891362318018039808", ctx.Value("ABC"))
					return NewEmptyResonse(), nil
				},
				func(ctx context.Context, r *http.Request) (HttpResponse, error) {
					logger.Infow(ctx, "http-get-03", "path", r.URL.Path)

					query := r.URL.Query()
					name := query.Get("name") // 获取 URL 参数

					reply := x.NewBuilder(x.WithKV("name", name)).Build()
					time.Sleep(500 * time.Millisecond)
					logger.Infow(ctx, "http-get", "path", r.URL.Path, "name", name)
					return NewStatusOkHttpResponse(reply, r.Header), nil
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
			c, err := NewHttpClient(ctx, logger, WithClientTimeout(3600))
			assert.Nil(t, err)
			r, err := c.Get(ctx, fmt.Sprintf("http://t.sunhe.org:%d%s", port, v.path), x.NewBuilder(x.WithKV("name", "pyro")), x.NewBuilder())
			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, r.StatusCode())
			assert.Equal(t, "application/json", r.Header().Get("Content-Type"))

			type httpReply struct {
				Code    int            `json:"code"`
				Message string         `json:"message"`
				Data    map[string]any `json:"data"`
			}

			reply := &httpReply{}
			err = json.Unmarshal(r.Body(), reply)
			logger.Infow(ctx, "http-reply", "reply", reply)
			assert.Nil(t, err)
			assert.Equal(t, x.NewBuilder(x.WithKV("name", "pyro")).Build(), reply.Data)
		}
		t.Run(n, f)
	}

	<-ctx.Done()
}

func Test_should(t *testing.T) {
	type Account struct {
		Name  string  `json:"name"      form:"name"`
		Age   int     `json:"age"       form:"age"`
		Score float32 `json:"score"     form:"score"`
	}

	var data = map[string]struct {
		input string
		exp   *Account
		ct    string
	}{
		"case-1": {
			input: `{"name":"puyu","age":99,"score":98.5}`,
			exp: &Account{
				Name:  "puyu",
				Age:   99,
				Score: 98.5,
			},
			ct: "application/json",
		},
		"case-2": {
			input: `name=puyu&age=99&score=98.5`,
			exp: &Account{
				Name:  "puyu",
				Age:   99,
				Score: 98.5,
			},
			ct: "application/x-www-form-urlencoded",
		},
	}

	for n, v := range data {
		f := func(t *testing.T) {
			act := &Account{}
			r := &http.Request{
				Method: http.MethodPost,
				Body:   io.NopCloser(strings.NewReader(v.input)),
				Header: http.Header{},
			}
			r.Header.Add("Content-Type", v.ct)
			err := ShouldBind(r, act)
			assert.Nil(t, err)
			assert.Equal(t, v.exp, act)
		}
		t.Run(n, f)
	}
}

func Test_should_form(t *testing.T) {
	type Account struct {
		Name  string  `form:"name"`
		Age   int     `form:"age"`
		Score float32 `form:"score"`
	}

	var data = map[string]struct {
		input url.Values
		exp   *Account
	}{
		"case-1": {
			input: url.Values{
				"name":  []string{"puyu"},
				"age":   []string{"99"},
				"score": []string{"98.5"},
			},
			exp: &Account{
				Name:  "puyu",
				Age:   99,
				Score: 98.5,
			},
		},
	}

	for n, v := range data {
		f := func(t *testing.T) {
			act := &Account{}
			r := &http.Request{
				Form: v.input,
			}
			err := ShouldBindForm(r, act)
			assert.Nil(t, err)
			assert.Equal(t, v.exp, act)
		}
		t.Run(n, f)
	}
}
