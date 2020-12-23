package netx

import (
	"context"
	"fmt"
	"testing"
	"time"
)

var tcpClientTestData = map[string]struct {
	host string
	port int
	pc   ProtocolConstructor
	msg  string
}{
	"case1": {
		host: "localhost",
		port: 6671,
		pc:   NewHBProtocol,
		msg:  "little package",
	},
}

func Test_tcpClient(t *testing.T) {
	for n, p := range tcpClientTestData {
		f := func(t *testing.T) {
			var c, err = NewTcpClient(&TcpClientOpt{
				Address: fmt.Sprintf("%s:%d", p.host, p.port),
				Timeout: time.Second,
				Retry:   3,
				PC:      p.pc,
			})
			if err != nil {
				t.Fatal(err)
				return
			}
			for {
				b, err := c.Send(context.TODO(), []byte(p.msg))
				if err != nil {
					//fmt.Println(err)
					continue
				}
				fmt.Println(string(b))
			}
		}
		t.Run(n, f)
	}
}
