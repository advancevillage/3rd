package netx

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type IHTTPCtx interface {
}

type httpCtx struct {
	engine *gin.Context
}

func newHTTPCtx(ctx *gin.Context) IHTTPCtx {
	return &httpCtx{engine: ctx}
}

type HTTPFunc func(IHTTPCtx)

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
}

type HTTPSrvOpt func(IHTTPServer)

func WithHTTPSrvAddr(h string, p int) HTTPSrvOpt {
	return func(s IHTTPServer) {
		s.addr(h, p)
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
	s.srv.ListenAndServe()
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

func (s *httpSrv) handle(method string, path string, f HTTPFunc) {
	handler := func(ctx *gin.Context) {
		var sctx = newHTTPCtx(ctx)
		f(sctx)
	}
	s.mux.Handle(method, path, handler)
}
