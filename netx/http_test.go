package netx

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"testing"
	"time"

	"github.com/advancevillage/3rd/logx"
	"github.com/advancevillage/3rd/mathx"
	"github.com/stretchr/testify/assert"
)

func init() {
	rand.Seed(time.Now().Unix())
}

type errorx struct {
	ErrorCode int    `json:"errorCode"`
	ErrorMsg  string `json:"errorMsg"`
}

var httpSrvCliTest = map[string]struct {
	host   string
	port   int
	method string
	path   string
	call   func(context.Context, IHTTPWriteReader)
}{
	"case-get": {
		host:   "127.0.0.1",
		port:   rand.Intn(1000) + 2048,
		method: http.MethodGet,
		path:   "/t/get",
		call: func(ctx context.Context, wr IHTTPWriteReader) {
			var traceId = wr.ReadParam(logx.TraceId)
			var l, err = logx.NewLogger("info")
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
	"case-post": {
		host:   "127.0.0.1",
		port:   rand.Intn(1000) + 2048,
		method: http.MethodPost,
		path:   "/t/post",
		call: func(ctx context.Context, wr IHTTPWriteReader) {
			var r = map[string]interface{}{}
			defer wr.Write(http.StatusOK, r)

			var errors []errorx
			r["errors"] = errors
			var body, err = wr.Read()
			if err != nil {
				errors = append(errors, errorx{ErrorCode: 1001, ErrorMsg: "read body"})
				return
			}
			type rq struct {
				TraceId string `json:"traceId"`
				Oth     string `json:"oth"`
			}

			var rr = new(rq)

			err = json.Unmarshal(body, rr)
			if err != nil {
				errors = append(errors, errorx{ErrorCode: 1002, ErrorMsg: "read body"})
				return
			}

			l, err := logx.NewLogger("info")
			if err != nil {
				errors = append(errors, errorx{ErrorCode: 1003, ErrorMsg: "read body"})
				return
			}
			l.Infow(ctx, "receive", "req", rr)
			r["statusCode"] = 0
			r["traceId"] = rr.TraceId
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
			//4. srv start
			go srv.Start()
			time.Sleep(2 * time.Second)
			hdr := map[string]string{
				"Content-Type": "application/json",
			}
			cli, err := NewHTTPCli(WithHTTPCliTimeout(3), WithHTTPCliHdr(hdr))
			if err != nil {
				t.Fatal(err)
				return
			}
			url := fmt.Sprintf("http://%s:%d%s", p.host, p.port, p.path)
			traceId := mathx.UUID()
			l.Infow(context.TODO(), "listen and serve", "host", p.host, "port", p.port, logx.TraceId, traceId)
			var b []byte
			switch p.method {
			case http.MethodGet:
				ps := map[string]string{
					logx.TraceId: traceId,
				}
				b, err = cli.GET(context.TODO(), url, ps, nil)
			case http.MethodPost:
				type rq struct {
					TraceId string `json:"traceId"`
					Oth     string `json:"oth"`
				}
				req := &rq{
					TraceId: traceId,
					Oth:     "oth",
				}
				buf, err := json.Marshal(req)
				if err != nil {
					t.Fatal(err)
					return
				}
				b, err = cli.POST(context.TODO(), url, nil, buf)
			default:
				t.Fatal("don't support method")
				return
			}
			if err != nil {
				t.Fatal(err)
				return
			}
			type reply struct {
				StatusCode int      `json:"statusCode"`
				TraceId    string   `json:"traceId"`
				Errors     []errorx `json:"errors"`
			}
			var rr = new(reply)

			err = json.Unmarshal(b, rr)
			if err != nil {
				t.Fatal(err)
				return
			}
			l.Infow(context.TODO(), "reply", "rr", rr)
			assert.Equal(t, traceId, rr.TraceId)
			time.Sleep(time.Second / 2)
		}
		t.Run(n, f)
	}
}
