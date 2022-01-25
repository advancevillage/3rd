package dht

import (
	"context"
	"testing"

	"github.com/advancevillage/3rd/logx"
)

var dhtTest = map[string]struct {
	network string
	addr    string
	level   string
	zone    uint16
}{
	"case1": {
		network: "udp",
		addr:    "192.168.187.176:5555",
		level:   "info",
		zone:    0x1010,
	},
}

func Test_dht(t *testing.T) {
	var ctx = context.TODO()
	for n, p := range dhtTest {
		f := func(t *testing.T) {
			var logger, err = logx.NewLogger(p.level)
			if err != nil {
				t.Fatal(err)
				return
			}
			s, err := NewDHT(ctx, logger, p.zone, p.network, p.addr)
			if err != nil {
				t.Fatal(err)
				return
			}
			s.Start()
		}
		t.Run(n, f)
	}
}
