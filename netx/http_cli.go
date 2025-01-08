package netx

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path/filepath"
	"time"

	"github.com/advancevillage/3rd/x"
)

type HttpResponse interface {
	Body() []byte
	Header() http.Header
	StatusCode() int
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

type IHTTPClient interface {
	GET(ctx context.Context, uri string, params x.Builder, headers x.Builder) (HttpResponse, error)
	POST(ctx context.Context, uri string, headers x.Builder, buf []byte) (HttpResponse, error)
	PostForm(ctx context.Context, uri string, params x.Builder, headers x.Builder) (HttpResponse, error)
	Upload(ctx context.Context, uri string, params x.Builder, headers x.Builder, field string, filename string, fieldReader io.Reader) (HttpResponse, error)

	hdr(h map[string]string)
	timeout(t int)
}

type HTTPCliOpt func(IHTTPClient)

var _ IHTTPClient = (*httpCli)(nil)

type httpCli struct {
	h  map[string]string
	tm int
}

func WithHTTPCliHdr(hdr map[string]string) HTTPCliOpt {
	return func(c IHTTPClient) {
		c.hdr(hdr)
	}
}

func WithHTTPCliTimeout(t int) HTTPCliOpt {
	return func(c IHTTPClient) {
		c.timeout(t)
	}
}

func NewHTTPCli(opts ...HTTPCliOpt) (IHTTPClient, error) {
	var cli = &httpCli{
		h:  make(map[string]string),
		tm: 3,
	}

	for _, opt := range opts {
		opt(cli)
	}

	return cli, nil
}

func (c *httpCli) hdr(h map[string]string) {
	c.h = h
}

func (c *httpCli) timeout(tm int) {
	if tm <= 0 {
		tm = 3
	}
	c.tm = tm
}

func (c *httpCli) GET(ctx context.Context, uri string, params x.Builder, headers x.Builder) (HttpResponse, error) {
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
	client = &http.Client{Timeout: time.Second * time.Duration(c.tm)}
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
	for k, v := range c.h {
		if _, ok := hdr[k]; ok {
			continue
		} else {
			request.Header.Add(k, v)
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

func (c *httpCli) POST(ctx context.Context, uri string, headers x.Builder, buf []byte) (HttpResponse, error) {
	var (
		client   *http.Client
		request  *http.Request
		response *http.Response
		err      error
		hdr      = headers.Build()
	)
	//1. 创建http客户端
	client = &http.Client{Timeout: time.Second * time.Duration(c.tm)}
	//2. 构造请求
	request, err = http.NewRequest(http.MethodPost, uri, bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}
	//3. 设置请求头 headers > config.Configure.Header
	for k, v := range hdr {
		request.Header.Add(k, fmt.Sprint(v))
	}
	for k, v := range c.h {
		if _, ok := hdr[k]; ok {
			continue
		} else {
			request.Header.Add(k, v)
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
	client = &http.Client{Timeout: time.Second * time.Duration(c.tm)}
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
	for k, v := range c.h {
		if _, ok := hdr[k]; ok {
			continue
		} else {
			request.Header.Add(k, v)
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
	client = &http.Client{Timeout: time.Second * time.Duration(c.tm)}
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
	for k, v := range c.h {
		if _, ok := hdr[k]; ok {
			continue
		} else {
			request.Header.Add(k, v)
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
