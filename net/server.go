//author: richard
package net

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/advancevillage/3rd/utils"

	"github.com/gin-gonic/gin"
)

type ModeType string

const (
	DebugMode   = ModeType(gin.DebugMode)
	TestMode    = ModeType(gin.TestMode)
	ReleaseMode = ModeType(gin.ReleaseMode)
)

type IHttpServer interface {
	StartServer()
	StopServer()
}

type IHttpContext interface {
	//@overview: 响应
	Write(code int, body interface{})

	ReadParam(q string) string
	ReadBody() ([]byte, error)

	ReadHeader(h string) string
	WriteHeader(headers map[string]string)

	ReadCookie(name string) (string, error)
	WriteCookie(name string, value string, path string, domain string, secure bool, httpOnly bool) error
}

type httpContext struct {
	engine *gin.Context
}

func newHttpContext(ctx *gin.Context) IHttpContext {
	return &httpContext{engine: ctx}
}

func (c *httpContext) ReadParam(q string) string {
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

func (c *httpContext) ReadBody() ([]byte, error) {
	var buf, err = ioutil.ReadAll(c.engine.Request.Body)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func (c *httpContext) ReadHeader(h string) string {
	return c.engine.GetHeader(h)
}

func (c *httpContext) WriteHeader(headers map[string]string) {
	for key := range headers {
		c.engine.Header(key, headers[key])
	}
}

func (c *httpContext) WriteCookie(name string, value string, path string, domain string, secure bool, httpOnly bool) error {
	var maxAge = 2 * 3600 //秒
	var cipherText, err = utils.EncryptUseAes([]byte(value))
	if err != nil {
		return err
	}
	text := base64.StdEncoding.EncodeToString(cipherText)
	c.engine.SetCookie(name, text, maxAge, path, domain, secure, httpOnly)
	return nil
}

func (c *httpContext) ReadCookie(name string) (string, error) {
	var value, err = c.engine.Cookie(name)
	if err != nil {
		return "", err
	}
	cipherText, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return "", err
	}
	plainText, err := utils.DecryptUseAes(cipherText)
	return string(plainText), err
}

func (c *httpContext) Write(code int, body interface{}) {
	c.engine.JSON(code, body)
}

type FuncHandler func(IHttpContext)

type IRouter interface {
	Add(method string, path string, f FuncHandler)
	Iterator(f func(method string, path string, f FuncHandler))
}

type router struct {
	method string
	path   string
	f      FuncHandler
}

type routeTable []*router

func NewRouter() IRouter {
	return &routeTable{}
}
func (c *routeTable) Add(method string, path string, f FuncHandler) {
	*c = append(*c, &router{method: method, path: path, f: f})
}

func (c *routeTable) Iterator(f func(method string, path string, f FuncHandler)) {
	for _, v := range *c {
		f(v.method, v.path, v.f)
	}
}

type httpServer struct {
	host   string
	port   int
	app    context.Context
	cancel context.CancelFunc
	rt     IRouter
	srv    *http.Server
	mux    *gin.Engine
}

func NewHttpServer(host string, port int, rt IRouter, m ModeType) IHttpServer {
	var s = httpServer{}
	s.host = host
	s.port = port
	s.rt = rt
	s.app, s.cancel = context.WithCancel(context.Background())
	gin.SetMode(string(m))
	s.mux = gin.New()
	s.srv = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", host, port),
		Handler: s.mux,
	}
	//init server
	s.initServer()
	return &s
}

func (s *httpServer) StartServer() {
	go s.start()
	go s.waitQuitSignal()
	select {
	case <-s.app.Done():
		s.close()
	}
}

func (s *httpServer) StopServer() {
	s.close()
}

func (s *httpServer) start() {
	var err = s.srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen %s:%d fail. %s\n", s.host, s.port, err.Error())
	}
}

func (s *httpServer) close() {
	var ctx, cancel = context.WithTimeout(context.TODO(), 3*time.Second)
	defer cancel()
	if err := s.srv.Shutdown(ctx); err != nil {
		log.Fatalf("server forced to shutdown: %s\n", err.Error())
	}
}

func (s *httpServer) initServer() {
	//init router
	s.rt.Iterator(s.handle)
}

func (s *httpServer) handle(method string, path string, f FuncHandler) {
	handler := func(ctx *gin.Context) {
		var hc = newHttpContext(ctx)
		f(hc)
	}
	s.mux.Handle(method, path, handler)
}

func (s *httpServer) waitQuitSignal() {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c
	s.cancel()
}
