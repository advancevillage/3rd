package dht

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/advancevillage/3rd/logx"
	"github.com/stretchr/testify/assert"
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

var dhtDHTHeap = map[string]struct {
	dist []uint8
	node []uint64
	k    uint8
	exp  []uint64
	err  error
}{
	"case1": {
		dist: []uint8{1, 2, 3, 4, 5, 6, 7, 7, 8},
		node: []uint64{100, 101, 102, 103, 104, 105, 106, 107, 108},
		k:    3,
		exp:  []uint64{100, 101, 102},
		err:  nil,
	},
	"case2": {
		dist: []uint8{1, 2, 3, 4, 5, 6, 7, 8},
		node: []uint64{100, 101, 102, 103, 104, 105, 106, 107, 108},
		k:    3,
		exp:  []uint64{100, 101, 102},
		err:  errInvalidParam,
	},
	"case3": {
		dist: []uint8{1, 2, 2, 3, 4, 5, 6, 7, 7, 8},
		node: []uint64{100, 101, 102, 103, 104, 105, 106, 107, 108, 109},
		k:    2,
		exp:  []uint64{100, 101},
		err:  nil,
	},
	"case4": {
		dist: []uint8{1, 1, 2, 3, 3, 4, 5, 6, 7, 7, 8},
		node: []uint64{100, 101, 102, 103, 104, 105, 106, 107, 108, 109, 99},
		k:    3,
		exp:  []uint64{100, 101, 102},
		err:  nil,
	},
	"case5": {
		dist: []uint8{1, 1, 2, 3, 3, 4, 5, 6, 7, 7, 8},
		node: []uint64{100, 101, 102, 103, 104, 105, 106, 107, 108, 109, 99},
		k:    11,
		exp:  []uint64{100, 101, 102, 103, 104, 105, 106, 107, 108, 109, 99},
		err:  nil,
	},
}

func Test_dht_heap(t *testing.T) {
	for n, p := range dhtDHTHeap {
		f := func(t *testing.T) {
			var h, err = NewDHTHeap(p.dist, p.node)
			if err != nil {
				assert.Equal(t, p.err, err)
				return
			}
			var act = h.Top(p.k)
			assert.Equal(t, p.exp, act)

		}
		t.Run(n, f)
	}
}
