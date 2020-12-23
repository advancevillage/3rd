package netx

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"testing"
	"time"

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
	host   string
	port   int
	ph     ProtocolHandler
	pc     ProtocolConstructor
	except interface{}
	err    error
}{
	"case1": {
		host: "localhost",
		port: rand.Intn(4096) + 4096,
		pc:   NewHBProtocol,
		ph: func(ctx context.Context, body []byte) ([]byte, error) {
			fmt.Println(string(body))
			return []byte("receive"), nil
		},
	},
}

func Test_TcpServer(t *testing.T) {
	for n, p := range tcpServerTestData {
		f := func(t *testing.T) {
			fmt.Println(p.host, p.port)
			var s, err = NewTcpServer(&TcpServerOpt{Host: p.host, Port: p.port, PC: p.pc, PH: p.ph})
			if err != nil {
				assert.Equal(t, err, p.err)
				return
			}
			s.StartServer()
		}
		t.Run(n, f)
	}
}
