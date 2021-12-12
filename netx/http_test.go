package netx

import (
	"context"
	"math/rand"
	"net/http"
	"testing"
	"time"

	"github.com/advancevillage/3rd/logx"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var httpSrvCliTest = map[string]struct {
	host   string
	port   int
	method string
	path   string
	call   func(context.Context, IHTTPWR)
}{
	"case-get": {
		host:   "0.0.0.0",
		port:   rand.Intn(1000) + 2048,
		method: http.MethodGet,
		path:   "/t/get",
		call: func(ctx context.Context, wr IHTTPWR) {
			var traceId = wr.ReadParam(logx.TraceId)
			var l, err = logx.NewLogger("info")
			type errorx struct {
				ErrorCode int    `json:"errorCode"`
				ErrorMsg  string `json:"errorMsg"`
			}
			var r = map[string]interface{}{}
			var errors []errorx
			if err != nil {
				errors = append(errors, errorx{ErrorCode: 1001, ErrorMsg: "create log fail"})
				r["statusCode"] = 1001
				r["errors"] = errors
			} else {
				l.Infow(ctx, "receive", logx.TraceId, traceId)
				r["statusCode"] = 0
			}
			r["traceId"] = traceId
			wr.Write(http.StatusOK, r)
		},
	},
}

func Test_srv_cli(t *testing.T) {
	for n, p := range httpSrvCliTest {
		f := func(t *testing.T) {
			//1. loger
			l, err := logx.NewLogger("info")
			if err != nil {
				t.Fatal(err)
				return
			}
			//2. route
			r := NewHTTPRouter()
			r.Add(p.method, p.path, p.call)
			//3. srv
			srv, err := NewHTTPSrv(WithHTTPSrvAddr(p.host, p.port), WithHTTPSrvRts(r), WithHTTPSrvLogger(l))
			if err != nil {
				t.Fatal(err)
				return
			}
			l.Infow(context.TODO(), "listen and serve", "host", p.host, "port", p.port)
			//4. srv start
			srv.Start()
		}
		t.Run(n, f)
	}
}
