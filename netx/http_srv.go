package netx

import (
	"context"
	"fmt"
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
	for i := 0; i < n; i++ {
		var (
			idx = i
			ff  = f[idx]
		)
		hf := func(c *gin.Context) {
			// 1. 设置上下文
			var (
				ctx       = c.Request.Context()
				trace, ok = c.Get(logx.TraceId)
			)
			if ok {
				ctx = context.WithValue(ctx, logx.TraceId, trace)
			}
			r, err := ff(ctx, c.Request)
			// 2. 系统错误
			if err != nil {
				r = NewInternalServerErrorHttpResponse(err)
				c.AbortWithStatusJSON(r.StatusCode(), r.Body())
				return
			}
			// 3. 设置响应头
			for k, v := range r.Header() {
				c.Header(k, v[0])
			}
			// 4. 提取Content-Type
			ct := r.Header().Get("Content-Type")
			if len(ct) <= 0 {
				ct = "application/json"
			}
			// 5. 中间件执行
			if idx < n-1 {
				c.Next()
				return
			}
			// 6. 设置耗时请求头
			c.Header(X_Request_Latency, fmt.Sprintf("%dms", time.Now().UnixNano()/1e6-c.GetInt64(X_Request_Latency)))
			// 7. 设置响应
			c.Data(r.StatusCode(), ct, r.Body())
		}
		fs = append(fs, hf)
	}
	s.srv.Handle(method, path, fs...)
}

func (s *httpSrv) withTraceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		trace, ok := c.Get(logx.TraceId)
		if !ok {
			trace = mathx.UUID()
			c.Set(logx.TraceId, trace)
		}
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
