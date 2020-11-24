//author: richard
package net

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"

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

type HttpContext struct {
	engine *gin.Context
}

func (c *HttpContext) ReadParam(q string) string {
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

func (c *HttpContext) ReadBody() ([]byte, error) {
	var buf, err = ioutil.ReadAll(c.engine.Request.Body)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func (c *HttpContext) ReadHeader(h string) string {
	return c.engine.GetHeader(h)
}

func (c *HttpContext) WriteHeader(headers map[string]string) {
	for key := range headers {
		c.engine.Header(key, headers[key])
	}
}

func (c *HttpContext) WriteCookie(name string, value string, path string, domain string, secure bool, httpOnly bool) error {
	var maxAge = 2 * 3600 //秒
	var cipherText, err = utils.EncryptUseAes([]byte(value))
	if err != nil {
		return err
	}
	text := base64.StdEncoding.EncodeToString(cipherText)
	c.engine.SetCookie(name, text, maxAge, path, domain, secure, httpOnly)
	return nil
}

func (c *HttpContext) ReadCookie(name string) (string, error) {
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

func (c *HttpContext) Write(code int, body interface{}) {
	c.engine.JSON(code, body)
}

type FuncHandler func(*HttpContext)

type router struct {
	method string
	path   string
	f      FuncHandler
}

type RouteTable []*router

func (c *RouteTable) Add(method string, path string, f FuncHandler) {
	*c = append(*c, &router{method: method, path: path, f: f})
}

type httpServer struct {
	host   string
	port   int
	app    context.Context
	cancel context.CancelFunc
	rt     RouteTable
	mux    *gin.Engine
}

func NewHttpServer(host string, port int, rt RouteTable, m ModeType) IHttpServer {
	var s = httpServer{}
	s.host = host
	s.port = port
	s.rt = rt
	gin.SetMode(string(m))
	s.mux = gin.New()
	//init server
	s.initServer()
	return &s
}

func (s *httpServer) StartServer() error {
	return s.mux.Run(fmt.Sprintf("%s:%d", s.host, s.port))
}

func (s *httpServer) initServer() {
	//init router
	for _, v := range s.rt {
		s.handle(v.method, v.path, v.f)
	}
}

func (s *httpServer) handle(method string, path string, f FuncHandler) {
	handler := func(ctx *gin.Context) {
		var hc = HttpContext{engine: ctx}
		f(&hc)
	}
	s.mux.Handle(method, path, handler)
}
