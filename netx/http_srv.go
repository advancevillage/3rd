package netx

import (
	"context"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/advancevillage/3rd/logx"
	"github.com/advancevillage/3rd/mathx"
	"github.com/gin-gonic/gin"
)

type httpRouter struct {
	method string
	path   string
	handle []HttpRegister
}

type httpSrv struct {
	opts   serverOptions
	logger logx.ILogger //日志

	srv     *gin.Engine        // https server
	rctx    context.Context    // root context
	rcancel context.CancelFunc // root cancel
}

func newHttpSrv(ctx context.Context, logger logx.ILogger, opt ...ServerOption) (*httpSrv, error) {
	// 0. 设置配置
	opts := defaultServerOptions
	for _, o := range opt {
		o.apply(&opts)
	}

	// 1. 服务对象
	s := &httpSrv{logger: logger, opts: opts}

	// 2. 上下文
	s.rctx, s.rcancel = context.WithCancel(ctx)

	// 3. 服务设置
	gin.SetMode(gin.ReleaseMode)
	s.srv = gin.New()

	// 4. 全局中间件
	s.srv.Use(s.withArrivalMiddleware(), s.withLatencyMiddleware(), s.withTraceMiddleware(), s.withNameMiddleware())

	// 5. 注册路由
	for _, r := range opts.rs {
		s.route(r.method, r.path, r.handle...)
	}

	return s, nil
}

func (s *httpSrv) Start() {
	go s.start()
	go waitQuitSignal(s.rcancel)
	<-s.rctx.Done()
	s.logger.Infow(s.rctx, "http server closed", "host", s.opts.host, "port", s.opts.port)
	time.Sleep(time.Second)
}

func (s *httpSrv) start() {
	s.logger.Infow(s.rctx, "https server start", "host", s.opts.host, "port", s.opts.port)
	err := s.srv.Run(fmt.Sprintf("%s:%d", s.opts.host, s.opts.port))
	if err != nil {
		s.logger.Errorw(s.rctx, "https server failed", "err", err, "host", s.opts.host, "port", s.opts.port)
	}
}

func (s *httpSrv) route(method, path string, f ...HttpRegister) {
	var (
		n  = len(f)
		fs = make([]gin.HandlerFunc, 0, n)
	)
	for i := range n {
		var (
			idx = i
			ff  = f[idx]
		)
		hf := func(c *gin.Context) {
			// 1. 设置上下文
			var ctx = s.createRequestContext(c.Request.Context(), c)
			var r, err = ff(ctx, c.Request)
			// 2. 系统错误
			if err != nil {
				c.Abort()
				r = NewInternalServerErrorHttpResponse(err)
				c.Data(r.StatusCode(), "application/json", r.Body())
				return
			}
			// 3. 设置响应头
			for k, v := range r.Header() {
				switch {
				case len(v) <= 0:
					// pass
				case strings.HasPrefix(k, rEQUEXT_CTX):
					s.updateRequestContext(c, strings.TrimLeft(k, rEQUEXT_CTX), strings.Join(v, ";"))

				default:
					c.Header(k, strings.Join(v, ";"))
				}
			}
			// 4. 提取Content-Type
			ct := r.Header().Get("Content-Type")
			if len(ct) <= 0 {
				ct = "application/json"
			}
			// 5. 设置耗时请求头
			c.Header(X_Request_Latency, fmt.Sprintf("%dms", time.Now().UnixNano()/1e6-c.GetInt64(X_Request_Latency)))
			// 6. 非200状态码
			if r.StatusCode() != http.StatusOK {
				c.Abort()
				c.Data(r.StatusCode(), ct, r.Body())
				return
			}
			// 7. 中间件执行
			if idx < n-1 {
				c.Next()
				return
			}
			// 8. 设置响应
			c.Data(r.StatusCode(), ct, r.Body())
		}
		fs = append(fs, hf)
	}
	s.srv.Handle(method, path, fs...)
}

func (s *httpSrv) withTraceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var (
			rctx  = url.Values{}
			trace = c.GetHeader(logx.TraceId)
		)
		if len(trace) <= 0 {
			trace = mathx.UUID()
		}
		rctx.Add(logx.UriId, c.Request.RequestURI)
		rctx.Add(logx.TraceId, trace)
		rctx.Add(logx.MethodId, c.Request.Method)
		c.Set(rEQUEXT_CTX, rctx.Encode())
		c.Header(logx.TraceId, fmt.Sprint(trace))
		c.Next()
	}
}

func (s *httpSrv) withLatencyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(X_Request_Latency, time.Now().UnixNano()/1e6)
	}
}

func (s *httpSrv) withArrivalMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		arrival := time.Now().UnixNano() / 1e6
		c.Header(X_Request_Arrival, fmt.Sprint(arrival))
		c.Next()
	}
}

func (s *httpSrv) withNameMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header(X_Request_Server, s.opts.name)
		c.Next()
	}
}

func (s *httpSrv) createRequestContext(ctx context.Context, c *gin.Context) context.Context {
	kv, err := url.ParseQuery(c.GetString(rEQUEXT_CTX))
	if err != nil {
		return ctx
	}
	for k, v := range kv {
		if len(v) <= 0 {
			continue
		}
		ctx = context.WithValue(ctx, k, strings.Join(v, ";"))
	}
	return ctx
}

func (s *httpSrv) updateRequestContext(c *gin.Context, k string, v string) {
	kv, err := url.ParseQuery(c.GetString(rEQUEXT_CTX))
	if err != nil {
		return
	}
	kk, err := hex.DecodeString(k)
	if err != nil {
		return
	}
	kv.Add(string(kk), v)
	c.Set(rEQUEXT_CTX, kv.Encode())
}
