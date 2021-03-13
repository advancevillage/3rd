package netx

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net"
	"testing"
	"time"

	"github.com/advancevillage/3rd/utils"
	"github.com/stretchr/testify/assert"
)

var connTestData = map[string]struct {
	maxSize uint32
	timeout time.Duration
	pk      *ecdsa.PrivateKey
	srt     *Secrets
}{
	"case1": {
		maxSize: 128,
		timeout: time.Minute,
	},
	"case2": {
		maxSize: 512,
		timeout: time.Minute,
	},
	"case3": {
		maxSize: 1024,
		timeout: time.Minute,
	},
	"case4": {
		maxSize: 2048,
		timeout: time.Minute,
	},
	"case5": {
		maxSize: 4096,
		timeout: time.Minute,
	},
	"case6": {
		maxSize: 8192,
		timeout: time.Minute,
	},
	"case7": {
		maxSize: 16384,
		timeout: time.Minute,
	},
	"case8": {
		maxSize: 32768,
		timeout: time.Minute,
	},
	"case9": {
		maxSize: 65536,
		timeout: time.Minute,
	},
	"case10": {
		maxSize: 131072,
		timeout: time.Minute,
	},
	"case11": {
		maxSize: 262144,
		timeout: time.Minute,
	},
	"case12": {
		maxSize: 524288,
		timeout: time.Minute,
	},
	"case13": {
		maxSize: 1048576,
		timeout: time.Minute,
	},
}

func Test_tcp_conn(t *testing.T) {
	for n, p := range connTestData {
		f := func(t *testing.T) {
			var s, c = net.Pipe()
			var err error
			p.pk, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
			if err != nil {
				t.Fatal(err)
				return
			}
			key, _ := hex.DecodeString("6368616e676520746869732070617373776f726420746f206120736563726574")
			p.srt = &Secrets{
				MK:      key,
				AK:      key,
				Egress:  sha256.New(),
				Ingress: sha256.New(),
			}
			var server ITransport
			var client ITransport
			server, err = NewConn(s, &TransportOption{MaxSize: p.maxSize, Timeout: p.timeout, PriKey: p.pk}, p.srt)
			if err != nil {
				t.Fatal(err)
				return
			}
			client, err = NewConn(c, &TransportOption{MaxSize: p.maxSize, Timeout: p.timeout, PriKey: p.pk}, p.srt)
			if err != nil {
				t.Fatal(err)
				return
			}
			for i := 0; i < 100; i++ {
				var im = utils.RandsString(int(p.maxSize))
				var rm []byte
				err = client.Write(context.TODO(), []byte(im))
				if err != nil {
					t.Fatal(err)
					return
				}
				rm, err = server.Read(context.TODO())
				if err != nil {
					t.Fatal(err)
					return
				}
				assert.Equal(t, im, string(rm))
				err = server.Write(context.TODO(), rm)
				if err != nil {
					t.Fatal(err)
					return
				}
				irm, err := client.Read(context.TODO())
				if err != nil {
					t.Fatal(err)
					return
				}
				assert.Equal(t, rm, irm)
			}
		}
		t.Run(n, f)
	}
}
