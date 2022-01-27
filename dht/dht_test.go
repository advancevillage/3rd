package dht

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/advancevillage/3rd/logx"
)

var dhtTest = map[string]struct {
	network string
	addr    string
	level   string
	zone    uint16
	seed    []uint64
}{
	"case1": {
		network: "udp",
		addr:    "192.168.187.176:5555",
		level:   "info",
		zone:    0x0000,
		seed:    []uint64{},
	},
}

func Test_dht(t *testing.T) {
	var ctx, cancel = context.WithCancel(context.TODO())
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		select {
		case <-ctx.Done():
		case <-c:
			cancel()
		}
		signal.Stop(c)
		close(c)
	}()

	for n, p := range dhtTest {
		f := func(t *testing.T) {
			var logger, err = logx.NewLogger(p.level)
			if err != nil {
				t.Fatal(err)
				return
			}
			s, err := NewDHT(ctx, logger, p.zone, p.network, p.addr, p.seed)
			if err != nil {
				t.Fatal(err)
				return
			}

			go s.Start()

			for {
				select {
				case <-ctx.Done():
					return
				case <-c:
					return
				default:
					time.Sleep(time.Second)
					logger.Infow(context.TODO(), "dump dht", "dump", s.Monitor())
				}
			}
		}
		t.Run(n, f)
	}
}
