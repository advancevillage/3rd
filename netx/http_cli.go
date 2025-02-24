package netx

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path/filepath"
	"time"

	"github.com/advancevillage/3rd/logx"
	"github.com/advancevillage/3rd/x"
)

func NewTooManyRequestsHttpResponse(err error) HttpResponse {
	return newCodeHttpResponse(http.StatusTooManyRequests, err.Error())
}

func NewBadRequestHttpResponse(err error) HttpResponse {
	return newCodeHttpResponse(http.StatusBadRequest, err.Error())
}

func NewUnauthorizedHttpResponse(err error) HttpResponse {
	return newCodeHttpResponse(http.StatusUnauthorized, err.Error())
}

func NewForbiddenHttpResponse(err error) HttpResponse {
	return newCodeHttpResponse(http.StatusForbidden, err.Error())
}

func NewNotFoundHttpResponse(err error) HttpResponse {
	return newCodeHttpResponse(http.StatusNotFound, err.Error())
}

func NewStatusOkHttpResponse(body []byte, hdr http.Header) HttpResponse {
	return newHttpResponse(body, hdr, http.StatusOK)
}

func NewInternalServerErrorHttpResponse(err error) HttpResponse {
	return newCodeHttpResponse(http.StatusInternalServerError, err.Error())
}

func newCodeHttpResponse(code int, message string) HttpResponse {
	b := x.NewBuilder(x.WithKV("httpCode", code), x.WithKV("httpMessage", message))
	body, err := json.Marshal(b.Build())
	if err != nil {
		return NewEmptyResonse()
	}
	return newHttpResponse(body, http.Header{}, code)
}

var _ HttpResponse = (*emptyHttpResponse)(nil)

type emptyHttpResponse struct {
	hdr http.Header
}

func newEmptyHttpResponse() *emptyHttpResponse {
	return &emptyHttpResponse{
		hdr: make(http.Header),
	}
}

func NewEmptyResonse() HttpResponse {
	return newEmptyHttpResponse()
}

func NewContextResponse(opt ...x.Option) HttpResponse {
	var (
		r = newEmptyHttpResponse()
		b = x.NewBuilder(opt...)
	)
	for k, v := range b.Build() {
		k = hex.EncodeToString([]byte(k))
		r.hdr.Add(fmt.Sprintf("%s%s", rEQUEXT_CTX, k), fmt.Sprint(v))
	}
	return r
}

func (c *emptyHttpResponse) Body() []byte {
	return []byte{}
}

func (c *emptyHttpResponse) Header() http.Header {
	return c.hdr
}

func (c *emptyHttpResponse) StatusCode() int {
	return http.StatusOK
}

var _ HttpResponse = (*httpResponse)(nil)

type httpResponse struct {
	body       []byte
	header     http.Header
	statusCode int
}

func newHttpResponse(body []byte, header http.Header, statusCode int) HttpResponse {
	return &httpResponse{
		body:       body,
		header:     header,
		statusCode: statusCode,
	}
}

func (c *httpResponse) Body() []byte {
	return c.body
}

func (c *httpResponse) Header() http.Header {
	return c.header
}

func (c *httpResponse) StatusCode() int {
	return c.statusCode
}

type HttpClient interface {
	Get(ctx context.Context, uri string, params x.Builder, headers x.Builder) (HttpResponse, error)
	Post(ctx context.Context, uri string, headers x.Builder, buf []byte) (HttpResponse, error)
	Upload(ctx context.Context, uri string, params x.Builder, headers x.Builder, field string, filename string, fieldReader io.Reader) (HttpResponse, error)
	PostForm(ctx context.Context, uri string, params x.Builder, headers x.Builder) (HttpResponse, error)
}

var _ HttpClient = (*httpCli)(nil)

type httpCli struct {
	opts   clientOptions
	logger logx.ILogger
}

func newHttpClient(ctx context.Context, logger logx.ILogger, opt ...ClientOption) (*httpCli, error) {
	opts := defaultClientOptions
	for _, o := range opt {
		o.apply(&opts)
	}
	return &httpCli{opts: opts, logger: logger}, nil
}

func (c *httpCli) Get(ctx context.Context, uri string, params x.Builder, headers x.Builder) (HttpResponse, error) {
	var (
		client   *http.Client
		request  *http.Request
		response *http.Response
		query    url.Values
		err      error
		q        = params.Build()
		hdr      = headers.Build()
	)
	//1. 创建http客户端
	client = &http.Client{Timeout: time.Second * time.Duration(c.opts.timeout)}
	//2. 构造请求
	request, err = http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}
	//3. 设置请求参数
	query = request.URL.Query()
	for k, v := range q {
		query.Add(k, fmt.Sprint(v))
	}
	request.URL.RawQuery = query.Encode()
	//4. 设置请求头 headers > config.Configure.Header
	for k, v := range hdr {
		request.Header.Add(k, fmt.Sprint(v))
	}
	for k, v := range c.opts.hdr {
		if _, ok := hdr[k]; ok {
			continue
		} else {
			request.Header.Add(k, fmt.Sprint(v))
		}
	}
	//5. 发送HTTP请求
	response, err = client.Do(request)
	if err != nil {
		return nil, err
	}
	//6. 请求结束后关闭连接
	defer response.Body.Close()
	//7. 读书响应数据
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	return newHttpResponse(body, response.Header, response.StatusCode), nil
}

func (c *httpCli) Post(ctx context.Context, uri string, headers x.Builder, buf []byte) (HttpResponse, error) {
	var (
		client   *http.Client
		request  *http.Request
		response *http.Response
		err      error
		hdr      = headers.Build()
	)
	//1. 创建http客户端
	client = &http.Client{Timeout: time.Second * time.Duration(c.opts.timeout)}
	//2. 构造请求
	request, err = http.NewRequest(http.MethodPost, uri, bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}
	//3. 设置请求头 headers > config.Configure.Header
	for k, v := range hdr {
		request.Header.Add(k, fmt.Sprint(v))
	}
	for k, v := range c.opts.hdr {
		if _, ok := hdr[k]; ok {
			continue
		} else {
			request.Header.Add(k, fmt.Sprint(v))
		}
	}
	//4. 发送HTTP请求
	response, err = client.Do(request)
	if err != nil {
		return nil, err
	}
	//5. 请求结束后关闭连接
	defer response.Body.Close()
	//6. 读书响应数据
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	return newHttpResponse(body, response.Header, response.StatusCode), nil
}

func (c *httpCli) PostForm(ctx context.Context, uri string, params x.Builder, headers x.Builder) (HttpResponse, error) {
	var (
		client   *http.Client
		request  *http.Request
		response *http.Response
		form     = url.Values{}
		err      error
		q        = params.Build()
		hdr      = headers.Build()
	)
	//1. 创建http客户端
	client = &http.Client{Timeout: time.Second * time.Duration(c.opts.timeout)}
	//2. 创建请求
	request, err = http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	//3. 设置From
	for k, v := range q {
		form.Add(k, fmt.Sprint(v))
	}
	//4. 设置请求头 headers > config.Configure.Header
	for k, v := range hdr {
		request.Header.Add(k, fmt.Sprint(v))
	}
	for k, v := range c.opts.hdr {
		if _, ok := hdr[k]; ok {
			continue
		} else {
			request.Header.Add(k, fmt.Sprint(v))
		}
	}
	//5. 发送HTTP请求
	response, err = client.Do(request)
	if err != nil {
		return nil, err
	}
	//6. 请求结束后关闭连接
	defer response.Body.Close()
	//7. 读书响应数据
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return newHttpResponse(body, response.Header, response.StatusCode), nil
}

func (c *httpCli) Upload(ctx context.Context, uri string, params x.Builder, headers x.Builder, field string, filename string, fieldReader io.Reader) (HttpResponse, error) {
	var (
		client   *http.Client
		request  *http.Request
		response *http.Response
		body     = &bytes.Buffer{}
		err      error
		q        = params.Build()
		hdr      = headers.Build()
	)
	//2. 分片
	writer := multipart.NewWriter(body)

	//2.1 文件上传
	if len(field) > 0 {
		part, err := writer.CreateFormFile(field, filepath.Base(filename))
		if err != nil {
			return nil, err
		}
		_, err = io.Copy(part, fieldReader)
		if err != nil {
			return nil, err
		}
	}

	//3. 设置参数
	for k, v := range q {
		err = writer.WriteField(k, fmt.Sprint(v))
		if err != nil {
			return nil, err
		}
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}
	//4. 创建HTTP客户端
	client = &http.Client{Timeout: time.Second * time.Duration(c.opts.timeout)}
	//5. 构建请求
	request, err = http.NewRequest(http.MethodPost, uri, body)
	if err != nil {
		return nil, err
	}
	//6. 设置请求头
	request.Header.Set("Content-Type", writer.FormDataContentType())
	for k, v := range hdr {
		request.Header.Add(k, fmt.Sprint(v))
	}
	for k, v := range c.opts.hdr {
		if _, ok := hdr[k]; ok {
			continue
		} else {
			request.Header.Add(k, fmt.Sprint(v))
		}
	}
	//7. 发送请求
	response, err = client.Do(request)
	if err != nil {
		return nil, err
	}
	//8. 关闭连接
	defer response.Body.Close()

	//9. 响应
	buf, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return newHttpResponse(buf, response.Header, response.StatusCode), nil
}
