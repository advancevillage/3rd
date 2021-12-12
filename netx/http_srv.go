package netx

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/advancevillage/3rd/logx"
	"github.com/gin-gonic/gin"
)

type IHTTPWR interface {
	Read() ([]byte, error)
	Write(code int, body interface{})

	ReadParam(q string) string
	WriteParam(params map[string]string)

	ReadHeader(h string) string
	WriteHeader(headers map[string]string)
}

type httpCtx struct {
	engine *gin.Context
}

func newHTTPCtx(ctx *gin.Context) IHTTPWR {
	return &httpCtx{engine: ctx}
}

func (c *httpCtx) Write(code int, body interface{}) {
	c.engine.JSON(code, body)
}

func (c *httpCtx) Read() ([]byte, error) {
	return ioutil.ReadAll(c.engine.Request.Body)
}

func (c *httpCtx) ReadParam(q string) string {
	var value = c.engine.PostForm(q)
	if len(value) <= 0 {
		value = c.engine.Query(q)
	}
	if len(value) <= 0 {
		value = c.engine.Param(q)
	}
	if len(value) <= 0 {
		value = c.engine.GetString(q)
	}
	return value
}

func (c *httpCtx) WriteParam(params map[string]string) {
	var qry = c.engine.Request.URL.Query()
	for k, v := range params {
		qry.Add(k, v)
	}
	c.engine.Request.URL.RawQuery = qry.Encode()
}

func (c *httpCtx) ReadHeader(h string) string {
	return c.engine.GetHeader(h)
}

func (c *httpCtx) WriteHeader(headers map[string]string) {
	for key := range headers {
		c.engine.Header(key, headers[key])
	}
}

type HTTPFunc func(context.Context, IHTTPWR)

type IHTTPRouter interface {
	Add(method string, path string, call HTTPFunc)
	iterator(f func(method string, path string, f HTTPFunc))
}

type rt struct {
	method string
	path   string
	call   HTTPFunc
}

type rts []*rt

func NewHTTPRouter() IHTTPRouter {
	return new(rts)
}

func (c *rts) Add(method string, path string, f HTTPFunc) {
	*c = append(*c, &rt{method: method, path: path, call: f})
}

func (c *rts) iterator(f func(method string, path string, call HTTPFunc)) {
	for _, v := range *c {
		f(v.method, v.path, v.call)
	}
}

type IHTTPServer interface {
	Start()

	addr(h string, p int)
	rts(IHTTPRouter)
	logger(l logx.ILogger)
}

type HTTPSrvOpt func(IHTTPServer)

func WithHTTPSrvAddr(h string, p int) HTTPSrvOpt {
	return func(s IHTTPServer) {
		s.addr(h, p)
	}
}

func WithHTTPSrvLogger(l logx.ILogger) HTTPSrvOpt {
	return func(s IHTTPServer) {
		s.logger(l)
	}
}

func WithHTTPSrvRts(rts IHTTPRouter) HTTPSrvOpt {
	return func(s IHTTPServer) {
		s.rts(rts)
	}
}

type httpSrv struct {
	host   string             //服务主机
	port   int                //服务端口
	ctx    context.Context    //上下文
	cancel context.CancelFunc //上下文取消函数
	r      IHTTPRouter        //路由
	srv    *http.Server       //HTTP服务
	mux    *gin.Engine        //HTTP服务引擎
	l      logx.ILogger       //日志
}

func NewHTTPSrv(opts ...HTTPSrvOpt) (IHTTPServer, error) {
	var s = &httpSrv{}
	s.ctx, s.cancel = context.WithCancel(context.Background())

	for _, opt := range opts {
		opt(s)
	}

	gin.SetMode(gin.DebugMode)
	s.mux = gin.New()

	s.srv = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", s.host, s.port),
		Handler: s.mux,
	}

	if s.r != nil {
		s.r.iterator(s.handle)
	}

	return s, nil
}

func (s *httpSrv) Start() {
	go s.start()
	go waitQuitSignal(s.cancel)
	select {
	case <-s.ctx.Done():
	}
}

func (s *httpSrv) start() {
	var err = s.srv.ListenAndServe()
	if err != nil {
		s.l.Errorw(s.ctx, "http server", "start", err)
	}
}

func (s *httpSrv) addr(h string, p int) {
	s.host = h
	s.port = p
}

func (s *httpSrv) rts(rts IHTTPRouter) {
	if rts == nil {
		s.r = NewHTTPRouter()
	} else {
		s.r = rts
	}
}

func (s *httpSrv) logger(l logx.ILogger) {
	s.l = l
}

func (s *httpSrv) handle(method string, path string, f HTTPFunc) {
	handler := func(ctx *gin.Context) {
		var wr = newHTTPCtx(ctx)
		var traceId = wr.ReadParam(logx.TraceId)
		var sctx = context.WithValue(ctx.Request.Context(), logx.TraceId, traceId)
		f(sctx, wr)
	}
	s.mux.Handle(method, path, handler)
}
