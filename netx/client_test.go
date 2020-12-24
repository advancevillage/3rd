package netx

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/advancevillage/3rd/utils"
)

var tcpClientTestData = map[string]struct {
	host string
	port int
	pc   ProtocolConstructor
	msg  int
}{
	//场景1: 小包
	"case1": {
		host: "localhost",
		port: 6671,
		pc:   NewHBProtocol,
		msg:  1500,
	},
}

func Test_tcpClient(t *testing.T) {
	for n, p := range tcpClientTestData {
		f := func(t *testing.T) {
			var c, err = NewTcpClient(&TcpClientOpt{
				Address: fmt.Sprintf("%s:%d", p.host, p.port),
				Timeout: time.Hour,
				Retry:   3,
				PC:      p.pc,
			})
			if err != nil {
				t.Fatal(err)
				return
			}
			var i = 0
			for {
				msg := fmt.Sprintf("%s:%d", utils.RandsString(p.msg), i)
				b, err := c.Send(context.TODO(), []byte(msg))
				if err != nil {
					continue
				}
				i++
				log.Println(string(b))
			}
		}
		t.Run(n, f)
	}
}
