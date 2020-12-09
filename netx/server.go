//author: richard
package netx

import (
	"bufio"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
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

var (
	ErrPartPackage = errors.New("receive part package")
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

type HttpFuncHandler func(IHttpContext)

type IRouter interface {
	Add(method string, path string, f HttpFuncHandler)
	Iterator(f func(method string, path string, f HttpFuncHandler))
}

type router struct {
	method string
	path   string
	f      HttpFuncHandler
}

type routeTable []*router

func NewRouter() IRouter {
	return &routeTable{}
}
func (c *routeTable) Add(method string, path string, f HttpFuncHandler) {
	*c = append(*c, &router{method: method, path: path, f: f})
}

func (c *routeTable) Iterator(f func(method string, path string, f HttpFuncHandler)) {
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
	go waitQuitSignal(s.cancel)
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

func (s *httpServer) handle(method string, path string, f HttpFuncHandler) {
	handler := func(ctx *gin.Context) {
		var hc = newHttpContext(ctx)
		f(hc)
	}
	s.mux.Handle(method, path, handler)
}

//@overview: 监听信号处理
func waitQuitSignal(cancel context.CancelFunc) {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c
	cancel()
}

//tcp request handler func
type ITcpWriter interface {
	Write([]byte) error
}

type tcpWriter struct {
	conn net.Conn
}

func newTcpWriter(conn net.Conn) ITcpWriter {
	return &tcpWriter{conn: conn}
}

func (c *tcpWriter) Write(b []byte) error {
	var _, err = c.conn.Write(b)
	return err
}

type TcpFuncHandler func(ITcpWriter, []byte)

type ITcpServer interface {
	StartServer()
}

//@overview: tcp server. 目标是更多请求更少的内存消耗
//@author: richard.sun
type tcpServer struct {
	host   string
	port   int
	app    context.Context
	cancel context.CancelFunc
	err    error
	ctt    time.Duration //conntection timeout
	p      IProtocol
}

func NewTcpServerWithProtocol(host string, port int, p IProtocol) (ITcpServer, error) {
	//1. 参数校验
	if port < 0 || port > 65535 {
		return nil, fmt.Errorf("port(%d) param is invalid", port)
	}
	var s = &tcpServer{}
	s.host = host
	s.port = port
	s.app, s.cancel = context.WithCancel(context.Background())
	s.p = p

	return s, nil
}

func (s *tcpServer) start() {
	var (
		err      error
		delay    time.Duration
		protocol = "tcp"
		addr     = fmt.Sprintf("%s:%d", s.host, s.port)
		l        net.Listener
		conn     net.Conn
	)
	//1. 监听端口
	l, err = net.Listen(protocol, addr)
	if err != nil {
		s.err = fmt.Errorf("listen %s fail. %s\n", addr, err.Error())
		s.cancel()
		return
	}
	//2. 监听请求
	for {
		conn, err = l.Accept()
		//accept fail
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				switch {
				case delay < 5*time.Millisecond:
					delay = 5 * time.Millisecond
				case delay < time.Second:
					delay = 2 * delay
				case delay >= time.Second:
					delay = time.Second
				}
				var t = time.NewTicker(delay)
				select {
				case <-t.C:
				case <-s.app.Done():
					t.Stop()
					return
				}
			}
			continue
		}
		delay = 0
		//处理
		go s.handler(conn)
	}
}

func (s *tcpServer) StartServer() {
	go s.start()
	go waitQuitSignal(s.cancel)
	select {
	case <-s.app.Done():
		s.close()
	}
}

func (s *tcpServer) handler(conn net.Conn) {
	//1. 关闭链接
	defer conn.Close()
	//2. 处理流
	var r = bufio.NewReader(conn)
	var w = bufio.NewWriter(conn)
	var err error
	var n int32
	var b []byte

	for {
		select {
		case <-s.app.Done():
			return
		default:
			//1. 解析协议Header
			n, err = s.p.ReadHeader(r)
			if err == io.EOF {
				return
			}
			if err != nil {
				continue
			}
			//2. 解析协议Body
			b, err = s.p.ReadBody(r, n)
			if err == io.EOF {
				return
			}
			if err != nil {
				continue
			}
			//3. 协议数据处理
			err = s.p.Write(w, b)
			if err == io.EOF {
				return
			}
		}
	}
}

func (s *tcpServer) close() {
	if s.err != nil {
		fmt.Println(s.err.Error())
	}
}
