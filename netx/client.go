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
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

var (
	errConnectClosed = errors.New("connection closed")
	errConnected     = errors.New("connection connected")
	errConnecting    = errors.New("connection connecting")
	errReconnected   = errors.New("reconnection")
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
	Send(context.Context, []byte) error
	Receive(context.Context) ([]byte, error)
}

type ClientOpt struct {
	Address string              // 服务端地址 ip:port
	Timeout time.Duration       // 客户端超时
	Retry   int                 //重试次数
	PC      ProtocolConstructor //协议生成器
}

//@overview: Tcp协议客户端应该具备超时、重试、断开重连、发送请求的基本功能
type tcpClient struct {
	cfg    *ClientOpt
	p      IProtocol
	conn   net.Conn
	app    context.Context
	cancel context.CancelFunc
	mu     sync.Mutex
	notify chan struct{}
	wg     sync.WaitGroup
}

func NewTcpClient(cfg *ClientOpt) (ITcpClient, error) {
	//1. 参数检查
	if cfg == nil || len(cfg.Address) <= 0 || cfg.Timeout <= 0 || cfg.Retry <= 0 || cfg.PC == nil {
		return nil, errors.New("tcp client options invalid param")
	}
	//2. 构建客户端
	var c = &tcpClient{}
	c.cfg = cfg
	c.app, c.cancel = context.WithCancel(context.Background())
	c.notify = make(chan struct{})
	go c.loop()
	return c, nil
}

func (c *tcpClient) loop() {
	var err error
	go c.halfOpen()
	for {
		err = c.connect(c.app)
		if err != nil {
			return
		}
		select {
		case <-c.notify:
			c.conn.Close()
		case <-c.app.Done():
			c.conn.Close()
			return
		}
	}
}

func (c *tcpClient) connect(ctx context.Context) error {
	for {
		select {
		case <-c.app.Done():
			return nil
		default:
			var conn, err = net.Dial("tcp", c.cfg.Address)
			//1. 连接失败 重试连接
			if err != nil {
				time.Sleep(50 * time.Millisecond)
				continue
			}
			//2. 连接成功
			c.mu.Lock()
			c.conn = conn
			c.p = c.cfg.PC(c.app, c.conn)
			c.mu.Unlock()
			go c.heartbeat()
			return nil
		}
	}
}

func (c *tcpClient) reconnect() {
	c.notify <- struct{}{}
}

func (c *tcpClient) heartbeat() {
	defer c.reconnect()
	var err error
	for {
		//1. 参数校验
		select {
		case <-c.app.Done():
			return
		default:
			err = c.send(c.app, c.p, nil)
			if err != nil {
				return
			}
			time.Sleep(time.Second / 5)
		}
	}
}

func (c *tcpClient) send(ctx context.Context, p IProtocol, body []byte) error {
	var err error
	if p == nil {
		return errConnecting
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-p.Done():
		return errReconnected
	default:
		err = p.Packet(ctx, body)
		if err != nil {
			return err
		}
		return nil
	}
}

func (c *tcpClient) receive(ctx context.Context, p IProtocol) ([]byte, error) {
	var (
		b   []byte
		err error
	)
	if p == nil {
		return nil, errConnecting
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-p.Done():
		return nil, errReconnected
	default:
		b, err = p.Unpacket(ctx)
		if err != nil {
			return nil, err
		}
		return b, nil
	}
}

func (c *tcpClient) Send(ctx context.Context, body []byte) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-c.app.Done():
		return c.app.Err()
	case <-c.notify:
		return errReconnected
	default:
		return c.send(ctx, c.p, body)
	}
}

func (c *tcpClient) Receive(ctx context.Context) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-c.app.Done():
		return nil, c.app.Err()
	case <-c.notify:
		return nil, errReconnected
	default:
		return c.receive(ctx, c.p)
	}
}

func (c *tcpClient) halfOpen() {
	var pp = make(chan os.Signal)
	signal.Notify(pp, syscall.SIGPIPE)
	select {
	case <-c.app.Done():
	case <-pp:
		c.reconnect()
	}
}

func (c *tcpClient) close() {
	c.cancel()
	time.Sleep(time.Second)
	close(c.notify)
}

//udp client
type IUdpClient interface {
	Send(context.Context, []byte) error
	Receive(context.Context) ([]byte, error)
}

type udpClient struct {
	cfg    *ClientOpt
	conn   *net.UDPConn
	app    context.Context
	cancel context.CancelFunc
	addr   *net.UDPAddr
	notify chan struct{}
}

func NewUdpClient(cfg *ClientOpt) (IUdpClient, error) {
	//1. 参数检查
	if cfg == nil || len(cfg.Address) <= 0 {
		return nil, errors.New("udp client options invalid param")
	}
	//2. 构建客户端
	var c = &udpClient{}
	var err error
	c.cfg = cfg
	c.app, c.cancel = context.WithCancel(context.Background())
	c.notify = make(chan struct{})
	c.addr, err = net.ResolveUDPAddr("udp", cfg.Address)
	if err != nil {
		return nil, err
	}
	c.conn, err = net.DialUDP("udp", nil, c.addr)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (c *udpClient) send(ctx context.Context, body []byte) error {
	var err error
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		_, err = c.conn.Write(body)
		if err != nil {
			return err
		}
		return nil
	}
}

func (c *udpClient) receive(ctx context.Context) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		var body = make([]byte, 2048)
		var n, err = c.conn.Read(body)
		if err != nil {
			return nil, err
		}
		return body[:n], nil
	}
}

func (c *udpClient) Send(ctx context.Context, body []byte) error {
	select {
	case <-c.app.Done():
		return c.app.Err()
	default:
		return c.send(ctx, body)
	}
}

func (c *udpClient) Receive(ctx context.Context) ([]byte, error) {
	select {
	case <-c.app.Done():
		return nil, c.app.Err()
	default:
		return c.receive(ctx)
	}
}
