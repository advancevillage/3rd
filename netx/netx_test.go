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
	host string
	port int
	ph   ProtocolHandler
	pc   ProtocolConstructor
	msg  int
}{
	//发送小报文
	"case1": {
		host: "localhost",
		pc:   NewStream,
		ph: func(ctx context.Context, body []byte) []byte {
			//fmt.Println("receive ", string(body))
			return body
		},
		msg: 16,
	},
	"case2": {
		host: "localhost",
		pc:   NewStream,
		ph: func(ctx context.Context, body []byte) []byte {
			//fmt.Println("receive ", string(body))
			return body
		},

		msg: 31,
	},
	"case3": {
		host: "localhost",
		pc:   NewStream,
		ph: func(ctx context.Context, body []byte) []byte {
			//fmt.Println("receive ", string(body))
			return body
		},
		msg: 63,
	},
	"case4": {
		host: "localhost",
		pc:   NewStream,
		ph: func(ctx context.Context, body []byte) []byte {
			//fmt.Println("receive ", string(body))
			return body
		},
		msg: 128,
	},
	"case5": {
		host: "localhost",
		pc:   NewStream,
		ph: func(ctx context.Context, body []byte) []byte {
			//fmt.Println("receive ", string(body))
			return body
		},
		msg: 255,
	},
	"case6": {
		host: "localhost",
		pc:   NewStream,
		ph: func(ctx context.Context, body []byte) []byte {
			//fmt.Println("receive ", string(body))
			return body
		},
		msg: 511,
	},
	"case7": {
		host: "localhost",
		pc:   NewStream,
		ph: func(ctx context.Context, body []byte) []byte {
			//fmt.Println("receive ", string(body))
			return body
		},
		msg: 1024,
	},
	"case8": {
		host: "localhost",
		pc:   NewStream,
		ph: func(ctx context.Context, body []byte) []byte {
			//fmt.Println("receive ", string(body))
			return body
		},
		msg: 2048,
	},
	"case9": {
		host: "localhost",
		pc:   NewStream,
		ph: func(ctx context.Context, body []byte) []byte {
			//fmt.Println("receive ", string(body))
			return body
		},
		msg: 4096,
	},
	"case10": {
		host: "localhost",
		pc:   NewStream,
		ph: func(ctx context.Context, body []byte) []byte {
			//fmt.Println("receive ", string(body))
			return body
		},
		msg: 8192,
	},
	"case11": {
		host: "localhost",
		pc:   NewStream,
		ph: func(ctx context.Context, body []byte) []byte {
			//fmt.Println("receive ", string(body))
			return body
		},
		msg: 16384,
	},
	"case12": {
		host: "localhost",
		pc:   NewStream,
		ph: func(ctx context.Context, body []byte) []byte {
			//fmt.Println("receive ", string(body))
			return body
		},
		msg: 32768,
	},
	"case13": {
		host: "localhost",
		pc:   NewStream,
		ph: func(ctx context.Context, body []byte) []byte {
			//fmt.Println("receive ", string(body))
			return body
		},
		msg: 65535,
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
			s, err = NewTcpServer(&TcpServerOpt{Host: p.host, Port: p.port, PC: p.pc, PH: p.ph})
			if err != nil {
				t.Fatal(err)
				return
			}
			//2. 构造客户端
			c, err = NewTcpClient(&TcpClientOpt{
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
	s, err = NewTcpServer(&TcpServerOpt{Host: host, Port: port, PC: NewStream, PH: ph})
	if err != nil {
		t.Fatal(err)
		return
	}
	s.StartServer()
}

func Test_OneTcpClient(t *testing.T) {
	var host = "localhost"
	var port = 8888
	var count = 1000
	log.Printf("%s:%d\n", host, port)
	var c ITcpClient
	var err error
	c, err = NewTcpClient(&TcpClientOpt{Address: fmt.Sprintf("%s:%d", host, port), Timeout: time.Hour, PC: NewStream, Retry: 3})
	if err != nil {
		t.Fatal(err)
		return
	}
	time.Sleep(5 * time.Second)
	go func() {
		for i := 0; i < count; i++ {
			var b, e = c.Receive(context.TODO())
			if e != nil {
				t.Fatal(e)
				return
			}
			log.Println("receive", string(b))
		}
	}()
	var wg sync.WaitGroup
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var msg = utils.RandsString(16)
			log.Println("send", msg)
			err = c.Send(context.TODO(), []byte(msg))
			if err != nil {
				t.Fatal(err)
				return
			}
		}()
	}
	wg.Wait()
	time.Sleep(time.Minute)
}
