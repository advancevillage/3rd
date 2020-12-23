package netx

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	errConnectClosed = errors.New("connection closed")
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

type ITcpClient interface {
	Send(context.Context, []byte) ([]byte, error)
}

type TcpClientOpt struct {
	Address string              // 服务端地址 ip:port
	Timeout time.Duration       // 客户端超时
	Retry   int                 //重试次数
	PC      ProtocolConstructor //协议生成器
}

//@overview: Tcp协议客户端应该具备超时、重试、断开重连、发送请求的基本功能
type tcpClient struct {
	cfg    *TcpClientOpt
	p      IProtocol
	conn   net.Conn
	app    context.Context
	cancel context.CancelFunc
	mu     sync.Mutex
	notify chan struct{}
}

func NewTcpClient(cfg *TcpClientOpt) (ITcpClient, error) {
	//1. 参数检查
	if cfg == nil || len(cfg.Address) <= 0 || cfg.Timeout <= 0 || cfg.Retry <= 0 || cfg.PC == nil {
		return nil, errors.New("tcp client options invalid param")
	}
	//2. 构建客户端
	var c = &tcpClient{}
	c.cfg = cfg
	c.app, c.cancel = context.WithCancel(context.Background())
	go c.loop()
	return c, nil
}

func (c *tcpClient) loop() {
	var err error
	for {
		err = c.conntect(c.app)
		if err != nil {
			return
		}
		select {
		case <-c.notify:
			c.clear()
		case <-c.app.Done():
			c.conn.Close()
			c.clear()
			return
		}
	}
}

func (c *tcpClient) conntect(ctx context.Context) error {
	fmt.Println(c.cfg.Address)
	for {
		var conn, err = net.DialTimeout("tcp", c.cfg.Address, c.cfg.Timeout)
		//1. 连接失败 重试连接
		if err != nil {
			fmt.Println("尝试重连")
			time.Sleep(50 * time.Millisecond)
			continue
		}
		//2. 连接失败
		if err != nil {
			return err
		}
		fmt.Println("连接成功")
		//3. 连接成功
		c.mu.Lock()
		c.conn = conn
		c.p = c.cfg.PC(c.conn)
		c.notify = make(chan struct{})
		c.mu.Unlock()
		go c.heartbeet(c.notify)
		break
	}
	return nil
}

func (c *tcpClient) clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.conn = nil
	c.p = nil
}

func (c *tcpClient) heartbeet(notify chan struct{}) {
	defer close(notify)
	var err error
	for {
		//1. 参数校验
		select {
		case <-c.app.Done():
			return
		default:
			if c.p == nil { //连接建立失败
				return
			}
			_, err = c.send(c.app, c.p, nil)
			if err != nil {
				return
			}
			time.Sleep(time.Second / 5)
		}
	}
}

func (c *tcpClient) send(ctx context.Context, p IProtocol, body []byte) ([]byte, error) {
	//0. 参数校验
	if p == nil {
		return nil, errConnectClosed
	}
	var (
		b   []byte
		err error
	)
	//1. 协议封包发送
	for i := 0; i < c.cfg.Retry; i++ {
		err = p.Packet(ctx, body)
		if err != nil {
			time.Sleep(50 * time.Millisecond)
			continue
		}
		break
	}
	err = p.Packet(ctx, body)
	if err != nil {
		return nil, err
	}
	//2. 协议接收解包
	b, err = p.Unpacket(ctx)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (c *tcpClient) Send(ctx context.Context, body []byte) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-c.app.Done():
		return nil, c.app.Err()
	default:
		return c.send(ctx, c.p, body)
	}
}
