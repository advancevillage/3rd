package netx

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
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
		pc:   NewHBProtocol,
		ph: func(ctx context.Context, body []byte) ([]byte, error) {
			//fmt.Println("receive ", string(body))
			return body, nil
		},
		msg: 16,
	},
	"case2": {
		host: "localhost",
		pc:   NewHBProtocol,
		ph: func(ctx context.Context, body []byte) ([]byte, error) {
			//fmt.Println("receive ", string(body))
			return body, nil
		},
		msg: 31,
	},
	"case3": {
		host: "localhost",
		pc:   NewHBProtocol,
		ph: func(ctx context.Context, body []byte) ([]byte, error) {
			//fmt.Println("receive ", string(body))
			return body, nil
		},
		msg: 63,
	},
	"case4": {
		host: "localhost",
		pc:   NewHBProtocol,
		ph: func(ctx context.Context, body []byte) ([]byte, error) {
			//fmt.Println("receive ", string(body))
			return body, nil
		},
		msg: 128,
	},
	"case5": {
		host: "localhost",
		pc:   NewHBProtocol,
		ph: func(ctx context.Context, body []byte) ([]byte, error) {
			//fmt.Println("receive ", string(body))
			return body, nil
		},
		msg: 255,
	},
	"case6": {
		host: "localhost",
		pc:   NewHBProtocol,
		ph: func(ctx context.Context, body []byte) ([]byte, error) {
			//fmt.Println("receive ", string(body))
			return body, nil
		},
		msg: 511,
	},
	"case7": {
		host: "localhost",
		pc:   NewHBProtocol,
		ph: func(ctx context.Context, body []byte) ([]byte, error) {
			//fmt.Println("receive ", string(body))
			return body, nil
		},
		msg: 1024,
	},
	"case8": {
		host: "localhost",
		pc:   NewHBProtocol,
		ph: func(ctx context.Context, body []byte) ([]byte, error) {
			//fmt.Println("receive ", string(body))
			return body, nil
		},
		msg: 2048,
	},
	"case9": {
		host: "localhost",
		pc:   NewHBProtocol,
		ph: func(ctx context.Context, body []byte) ([]byte, error) {
			//fmt.Println("receive ", string(body))
			return body, nil
		},
		msg: 4096,
	},
	"case10": {
		host: "localhost",
		pc:   NewHBProtocol,
		ph: func(ctx context.Context, body []byte) ([]byte, error) {
			//fmt.Println("receive ", string(body))
			return body, nil
		},
		msg: 8192,
	},
	"case11": {
		host: "localhost",
		pc:   NewHBProtocol,
		ph: func(ctx context.Context, body []byte) ([]byte, error) {
			//fmt.Println("receive ", string(body))
			return body, nil
		},
		msg: 16384,
	},
	"case12": {
		host: "localhost",
		pc:   NewHBProtocol,
		ph: func(ctx context.Context, body []byte) ([]byte, error) {
			//fmt.Println("receive ", string(body))
			return body, nil
		},
		msg: 32768,
	},
	"case13": {
		host: "localhost",
		pc:   NewHBProtocol,
		ph: func(ctx context.Context, body []byte) ([]byte, error) {
			//fmt.Println("receive ", string(body))
			return body, nil
		},
		msg: 65535,
	},
}

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
			var table = make(map[string]struct{})
			var i = 0
			var b []byte
			for {
				select {
				case <-tg.C:
					if i <= 0 {
						t.Fatal("unit test fail")
					}
					time.Sleep(time.Second * 5)
					if len(table) > 0 {
						t.Fatal("table has pkg data", len(table))
					}
					c.Close()
					s.StopServer()
					fmt.Printf("unit pass %d samples. fail %d\n", i, len(table))
					return
				default:
					var msg = fmt.Sprintf("%s:%d", utils.RandsString(p.msg), i)
					table[msg] = struct{}{}
					//4. 服务端接收请求
					b, err = c.Send(context.TODO(), []byte(msg))
					if err != nil {
						fmt.Println(err)
						delete(table, msg)
						continue
					}
					i++
					switch n {
					case "case9", "case10", "case11", "case12", "case13":
						//fmt.Println(i)
					}
					//5. 验证传输数据的正确性
					if _, ok := table[string(b)]; ok {
						delete(table, string(b))
					}
				}
			}
		}
		t.Run(n, f)
	}
}
