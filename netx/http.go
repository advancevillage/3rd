//author: richard
package netx

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
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

/////////////////////////////////////////////////////////////////////////////////////
//http client
type IHttpClient interface {
	GET(ctx context.Context, uri string, params map[string]string, headers map[string]string) ([]byte, error)
	POST(ctx context.Context, uri string, headers map[string]string, buf []byte) ([]byte, error)
	PostForm(ctx context.Context, uri string, params map[string]string, headers map[string]string) ([]byte, error)
	Upload(ctx context.Context, uri string, params map[string]string, headers map[string]string, upload string) ([]byte, error)
}

//@overview: 封装HTTP客户端库
type httpClient struct {
	headers map[string]string
	timeout int64
	retry   uint
}

func NewHttpClient(headers map[string]string, timeout int64, retry uint) (IHttpClient, error) {
	if retry <= 0 {
		return nil, fmt.Errorf("invalid retry=%d param setting", retry)
	}
	if timeout <= 0 {
		return nil, fmt.Errorf("invalid timeout=%d param setting", timeout)
	}
	return &httpClient{
		headers: headers,
		timeout: timeout,
		retry:   retry,
	}, nil
}

func (c *httpClient) GET(ctx context.Context, uri string, params map[string]string, headers map[string]string) ([]byte, error) {
	var (
		client   *http.Client
		request  *http.Request
		response *http.Response
		query    url.Values
		err      error
		t        = time.NewTicker(50 * time.Millisecond)
	)
	defer t.Stop()
	//1. 创建http客户端
	client = &http.Client{Timeout: time.Second * time.Duration(c.timeout)}
	//2. 构造请求
	request, err = http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}
	//3. 设置请求参数
	query = request.URL.Query()
	for k, v := range params {
		query.Add(k, v)
	}
	request.URL.RawQuery = query.Encode()
	//4. 设置请求头 headers > config.Configure.Header
	for k, v := range headers {
		request.Header.Add(k, v)
	}
	for k, v := range c.headers {
		if _, ok := headers[k]; ok {
			continue
		} else {
			request.Header.Add(k, v)
		}
	}
	//5. 发送HTTP请求
	for i := uint(0); i < c.retry; i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			response, err = client.Do(request)
		}
		if err == nil {
			break
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-t.C:
		}
	}
	if err != nil {
		return nil, err
	}
	//6. 请求结束后关闭连接
	defer response.Body.Close()
	//7. 读书响应数据
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (c *httpClient) POST(ctx context.Context, uri string, headers map[string]string, buf []byte) ([]byte, error) {
	var (
		client   *http.Client
		request  *http.Request
		response *http.Response
		err      error
		t        = time.NewTicker(50 * time.Millisecond)
	)
	defer t.Stop()
	//1. 创建http客户端
	client = &http.Client{Timeout: time.Second * time.Duration(c.timeout)}
	//2. 构造请求
	request, err = http.NewRequest(http.MethodPost, uri, bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}
	//3. 设置请求头 headers > config.Configure.Header
	for k, v := range headers {
		request.Header.Add(k, v)
	}
	for k, v := range c.headers {
		if _, ok := headers[k]; ok {
			continue
		} else {
			request.Header.Add(k, v)
		}
	}
	//4. 发送HTTP请求
	for i := uint(0); i < c.retry; i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			response, err = client.Do(request)
		}
		if err == nil {
			break
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-t.C:
		}
	}
	if err != nil {
		return nil, err
	}
	//5. 请求结束后关闭连接
	defer response.Body.Close()
	//6. 读书响应数据
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (c *httpClient) PostForm(ctx context.Context, uri string, params map[string]string, headers map[string]string) ([]byte, error) {
	var (
		client   *http.Client
		request  *http.Request
		response *http.Response
		form     = url.Values{}
		err      error
		t        = time.NewTicker(50 * time.Millisecond)
	)
	defer t.Stop()
	//1. 创建http客户端
	client = &http.Client{Timeout: time.Second * time.Duration(c.timeout)}
	//2. 创建请求
	request, err = http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	//3. 设置From
	for k, v := range params {
		form.Add(k, v)
	}
	//4. 设置请求头 headers > config.Configure.Header
	for k, v := range headers {
		request.Header.Add(k, v)
	}
	for k, v := range c.headers {
		if _, ok := headers[k]; ok {
			continue
		} else {
			request.Header.Add(k, v)
		}
	}
	//5. 发送HTTP请求
	for i := uint(0); i < c.retry; i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			response, err = client.Do(request)
		}
		if err == nil {
			break
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-t.C:
		}
	}
	if err != nil {
		return nil, err
	}
	//6. 请求结束后关闭连接
	defer response.Body.Close()
	//7. 读书响应数据
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (c *httpClient) Upload(ctx context.Context, uri string, params map[string]string, headers map[string]string, upload string) ([]byte, error) {
	var (
		client   *http.Client
		request  *http.Request
		response *http.Response
		file     *os.File
		body     = &bytes.Buffer{}
		err      error
		t        = time.NewTicker(50 * time.Millisecond)
	)
	defer t.Stop()
	//1. 读取文件
	file, err = os.Open(upload)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	//2. 分片
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(upload))
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return nil, err
	}
	//3. 设置参数
	for k, v := range params {
		err = writer.WriteField(k, v)
		if err != nil {
			return nil, err
		}
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}
	//4. 创建HTTP客户端
	client = &http.Client{Timeout: time.Second * time.Duration(c.timeout)}
	//5. 构建请求
	request, err = http.NewRequest(http.MethodPost, uri, body)
	if err != nil {
		return nil, err
	}
	//6. 设置请求头
	request.Header.Set("Content-Type", writer.FormDataContentType())
	for k, v := range headers {
		request.Header.Add(k, v)
	}
	for k, v := range c.headers {
		if _, ok := headers[k]; ok {
			continue
		} else {
			request.Header.Add(k, v)
		}
	}
	//7. 发送请求
	for i := uint(0); i < c.retry; i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			response, err = client.Do(request)
		}
		if err == nil {
			break
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-t.C:
		}
	}
	if err != nil {
		return nil, err
	}
	//8. 关闭连接
	defer response.Body.Close()
	//9. 响应
	buf, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	return buf, nil
}
