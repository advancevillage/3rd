package netx

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/advancevillage/3rd/utils"
	"github.com/stretchr/testify/assert"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var httpServerTestData = map[string]struct {
	rs      IRouter
	host    string
	port    int
	method  string
	path    string
	params  map[string]string
	handler HttpFuncHandler
	except  string
}{
	//测试Get请求并携带Query
	"case1": {
		host:   "localhost",
		port:   rand.Intn(4096) + 4096,
		rs:     NewRouter(),
		method: http.MethodGet,
		path:   "test/aaa",
		params: map[string]string{
			"q": "1111",
		},
		handler: func(hc IHttpContext) {
			var q = hc.ReadParam("q")
			var v, _ = strconv.Atoi(q)
			hc.Write(http.StatusOK, v)
		},
		except: "1111",
	},
}

func Test_HttpServer(t *testing.T) {
	for n, p := range httpServerTestData {
		f := func(t *testing.T) {
			p.rs.Add(p.method, p.path, p.handler)
			var s = NewHttpServer(p.host, p.port, p.rs, DebugMode)
			//1. Http Server
			go s.StartServer()
			//2. Http Client
			c, err := NewHttpClient(nil, 2, 1)
			if err != nil {
				t.Fatal(err)
				return
			}
			time.Sleep(time.Second)
			buf, err := c.GET(context.TODO(), fmt.Sprintf("http://%s:%d/%s", p.host, p.port, p.path), p.params, nil)
			if err != nil {
				t.Fatal(err.Error())
			}
			s.StopServer()
			assert.Equal(t, p.except, string(buf))
		}
		t.Run(n, f)
	}
}

//tcp unit test
var tcpServerTestData = map[string]struct {
	host  string
	port  int
	ph    ProtocolHandler
	pc    ProtocolConstructor
	msg   int
	count int
}{
	//发送小报文
	"case1": {
		host: "localhost",
		pc:   NewStream,
		ph: func(ctx context.Context, body []byte) []byte {
			//fmt.Println("receive ", string(body))
			return body
		},
		msg:   16,
		count: 3000,
	},
	"case2": {
		host: "localhost",
		pc:   NewStream,
		ph: func(ctx context.Context, body []byte) []byte {
			//fmt.Println("receive ", string(body))
			return body
		},
		count: 3000,
		msg:   31,
	},
	"case3": {
		host: "localhost",
		pc:   NewStream,
		ph: func(ctx context.Context, body []byte) []byte {
			//fmt.Println("receive ", string(body))
			return body
		},
		msg:   63,
		count: 3000,
	},
	"case4": {
		host: "localhost",
		pc:   NewStream,
		ph: func(ctx context.Context, body []byte) []byte {
			//fmt.Println("receive ", string(body))
			return body
		},
		msg:   128,
		count: 3000,
	},
	"case5": {
		host: "localhost",
		pc:   NewStream,
		ph: func(ctx context.Context, body []byte) []byte {
			//fmt.Println("receive ", string(body))
			return body
		},
		msg:   255,
		count: 3000,
	},
	"case6": {
		host: "localhost",
		pc:   NewStream,
		ph: func(ctx context.Context, body []byte) []byte {
			//fmt.Println("receive ", string(body))
			return body
		},
		msg:   511,
		count: 3000,
	},
	"case7": {
		host: "localhost",
		pc:   NewStream,
		ph: func(ctx context.Context, body []byte) []byte {
			//fmt.Println("receive ", string(body))
			return body
		},
		msg:   1024,
		count: 3000,
	},
	"case8": {
		host: "localhost",
		pc:   NewStream,
		ph: func(ctx context.Context, body []byte) []byte {
			//fmt.Println("receive ", string(body))
			return body
		},
		msg:   2048,
		count: 3000,
	},
	"case9": {
		host: "localhost",
		pc:   NewStream,
		ph: func(ctx context.Context, body []byte) []byte {
			//fmt.Println("receive ", string(body))
			return body
		},
		msg:   4096,
		count: 3000,
	},
	"case10": {
		host: "localhost",
		pc:   NewStream,
		ph: func(ctx context.Context, body []byte) []byte {
			//fmt.Println("receive ", string(body))
			return body
		},
		msg:   8192,
		count: 3000,
	},
	"case11": {
		host: "localhost",
		pc:   NewStream,
		ph: func(ctx context.Context, body []byte) []byte {
			//fmt.Println("receive ", string(body))
			return body
		},
		msg:   16384,
		count: 3000,
	},
	"case12": {
		host: "localhost",
		pc:   NewStream,
		ph: func(ctx context.Context, body []byte) []byte {
			//fmt.Println("receive ", string(body))
			return body
		},
		msg:   32768,
		count: 3000,
	},
	"case13": {
		host: "localhost",
		pc:   NewStream,
		ph: func(ctx context.Context, body []byte) []byte {
			//fmt.Println("receive ", string(body))
			return body
		},
		msg:   65535,
		count: 3000,
	},
}

//go test -v -count=1 -timeout 20m -test.run Test_TcpServer .
func Test_TcpServer(t *testing.T) {
	rand.Seed(time.Now().Unix())
	for n, p := range tcpServerTestData {
		f := func(t *testing.T) {
			p.port = rand.Intn(4096) + 8192
			fmt.Println(p.host, p.port, time.Now().Unix(), p.msg)
			var s ITcpServer
			var c ITcpClient
			var err error
			//1. 构造服务端
			s, err = NewTcpServer(&ServerOpt{Host: p.host, Port: p.port, PC: p.pc, PH: p.ph})
			if err != nil {
				t.Fatal(err)
				return
			}
			//2. 构造客户端
			c, err = NewTcpClient(&ClientOpt{
				Address: fmt.Sprintf("%s:%d", p.host, p.port),
				Timeout: time.Hour,
				Retry:   3,
				PC:      p.pc,
			})
			if err != nil {
				t.Fatal(err)
				return
			}
			//3. 启动服务端
			go s.StartServer()
			time.Sleep(5 * time.Second)
			//4. 客户端发送请求
			var tg = time.NewTicker(time.Second * 60)
			var table = map[string]struct{}{}
			var count int
			//5. 接受服务端数据流
			for {
				select {
				case <-tg.C:
					time.Sleep(time.Second * 5)
					fmt.Println(n, "pass", count)
					s.StopServer()
					assert.Equal(t, 0, len(table))
					return
				default:
					var msg = utils.RandsString(p.msg)
					table[msg] = struct{}{}
					//4. 服务端接收请求
					err = c.Send(context.TODO(), []byte(msg))
					if err != nil {
						delete(table, msg)
						continue
					}
					count++
					var b, e = c.Receive(context.TODO())
					if e != nil {
						t.Fatal(e)
						return
					}
					//验证传输数据的正确性
					if _, ok := table[string(b)]; ok {
						delete(table, string(b))
					}
				}
			}
		}
		t.Run(n, f)
	}
}

func Test_OneTcpServer(t *testing.T) {
	var host = "localhost"
	var port = 8888
	var ph = func(ctx context.Context, body []byte) []byte {
		return body
	}
	log.Printf("%s:%d\n", host, port)
	var s ITcpServer
	var err error
	s, err = NewSIPServer(&ServerOpt{Host: host, Port: port, PC: NewStream, PH: ph})
	if err != nil {
		t.Fatal(err)
		return
	}
	s.StartServer()
}

func Test_OneTcpClient(t *testing.T) {
	var host = "localhost"
	var port = 8888
	var c ISIPClient
	var err error
	fmt.Printf("%s:%d\n", host, port)
	c, err = NewSIPClent(&ClientOpt{Address: fmt.Sprintf("%s:%d", host, port), Timeout: time.Second * 10, PC: NewStream, Retry: 3})
	if err != nil {
		t.Fatal(err)
		return
	}
	time.Sleep(5 * time.Second)
	var wg sync.WaitGroup
	var count = 3000
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			var msg = utils.RandsString(12)
			//fmt.Printf("send msg(%d) %s\n", index, msg)
			var b, e = c.Send(context.TODO(), []byte(msg))
			if e != nil {
				t.Fatal(e)
				return
			}
			//fmt.Printf("receive msg(%d) %s\n", index, string(b))
			assert.Equal(t, string(b), msg)
		}(i)
	}
	wg.Wait()
}

func Test_TcpSIPServer(t *testing.T) {
	rand.Seed(time.Now().Unix())
	for n, p := range tcpServerTestData {
		f := func(t *testing.T) {
			p.port = rand.Intn(4096) + 8192
			fmt.Println(p.host, p.port, time.Now().Unix(), p.msg)
			var s ISIPServer
			var c ISIPClient
			var err error
			//1. 构造服务端
			s, err = NewSIPServer(&ServerOpt{Host: p.host, Port: p.port, PC: p.pc, PH: p.ph})
			if err != nil {
				t.Fatal(err)
				return
			}
			//2. 构造客户端
			c, err = NewSIPClent(&ClientOpt{
				Address: fmt.Sprintf("%s:%d", p.host, p.port),
				Timeout: time.Second * 10,
				Retry:   3,
				PC:      p.pc,
			})
			if err != nil {
				t.Fatal(err)
				return
			}
			//3. 启动服务端
			go s.StartServer()
			time.Sleep(5 * time.Second)
			//4. 客户端发送请求
			var wg sync.WaitGroup
			for i := 0; i < p.count; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					var msg = utils.RandsString(p.msg)
					//4. 服务端接收请求
					var b, e = c.Send(context.TODO(), []byte(msg))
					if err != nil {
						t.Fatal(e)
						return
					}
					//验证传输数据的正确性
					assert.Equal(t, string(b), msg)
				}()
			}
			wg.Wait()
			s.StopServer()
		}
		t.Run(n, f)
	}
}

//udp server
func Test_OneUdpServer(t *testing.T) {
	var host = "192.168.1.101"
	var port = 8888
	var ph = func(ctx context.Context, body []byte) []byte {
		return body
	}
	var s IUdpServer
	var err error
	s, err = NewUdpServer(&ServerOpt{Host: host, Port: port, PC: NewStream, PH: ph})
	if err != nil {
		t.Fatal(err)
		return
	}
	s.StartServer()
}

func Test_OneUdpClient(t *testing.T) {
	var host = "192.168.1.101"
	var port = 8888
	var c IUdpClient
	var err error
	c, err = NewUdpClient(&ClientOpt{Address: fmt.Sprintf("%s:%d", host, port), Timeout: 10})
	if err != nil {
		t.Fatal(err)
		return
	}
	log.Println(c)
	time.Sleep(time.Minute)
}

var udpCSTestData = map[string]struct {
	host  string
	port  int
	size  int
	count int
	ms    int
}{
	"case1": {
		host:  "127.0.0.1",
		size:  10,
		count: 10000,
		ms:    1500,
	},
	"case2": {
		host:  "127.0.0.1",
		size:  20,
		count: 10000,
		ms:    1500,
	},
	"case3": {
		host:  "127.0.0.1",
		size:  40,
		count: 10000,
		ms:    1500,
	},
	"case4": {
		host:  "127.0.0.1",
		size:  80,
		count: 10000,
		ms:    1500,
	},
	"case5": {
		host:  "127.0.0.1",
		size:  160,
		count: 10000,
		ms:    1500,
	},
	"case6": {
		host:  "127.0.0.1",
		size:  320,
		count: 10000,
		ms:    1500,
	},
	"case7": {
		host:  "127.0.0.1",
		size:  640,
		count: 10000,
		ms:    1500,
	},
	"case8": {
		host:  "127.0.0.1",
		size:  1280,
		count: 10000,
		ms:    1500,
	},
	"case9": {
		host:  "127.0.0.1",
		size:  2560,
		count: 10000,
		ms:    3200,
	},
	"case10": {
		host:  "127.0.0.1",
		size:  5120,
		count: 10000,
		ms:    6400,
	},
}

func Test_UdpCS(t *testing.T) {
	rand.Seed(time.Now().Unix())
	for n, p := range udpCSTestData {
		f := func(t *testing.T) {
			p.port = rand.Intn(4096) + 4096
			var c IUdpClient
			var s IUdpServer
			var err error
			var sm, rm = make(map[string]struct{}), make(map[string]struct{})
			var ph = func(ctx context.Context, body []byte) []byte {
				return body
			}
			s, err = NewUdpServer(&ServerOpt{Host: p.host, Port: p.port, PH: ph, MaxSize: p.ms})
			if err != nil {
				t.Error(err)
				return
			}
			c, err = NewUdpClient(&ClientOpt{Address: fmt.Sprintf("%s:%d", p.host, p.port), Timeout: 10 * time.Second, MaxSize: p.ms})
			if err != nil {
				t.Error(err)
				return
			}
			//1. 启动服务端
			go s.StartServer()
			time.Sleep(time.Second * 3)
			//2. 启动客户端发送
			go func() {
				for i := 0; i < p.count; i++ {
					var msg = utils.RandsString(p.size)
					sm[msg] = struct{}{}
					var e = c.Send(context.TODO(), []byte(msg))
					if e != nil {
						fmt.Println(err.Error())
						t.Error(e)
						return
					}
				}
				time.Sleep(time.Second * 2)
			}()
			//3. 启动客户端接收
			var ctx, cancel = context.WithTimeout(context.TODO(), time.Second*10)
			defer cancel()
			for {
				var msg, err = c.Receive(ctx)
				if err != nil {
					break
				}
				rm[string(msg)] = struct{}{}
			}
			s.StopServer()
			//4. 校验
			for k := range sm {
				if _, ok := rm[k]; ok {
					delete(sm, k)
					delete(rm, k)
				}
			}
			if len(sm) <= 0 && len(rm) <= 0 {
				t.Log("success")
			} else {
				t.Error("fail")
			}
		}
		t.Run(n, f)
	}
}