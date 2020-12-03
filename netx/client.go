package netx

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

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
