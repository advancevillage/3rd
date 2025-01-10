package netx

import (
	"context"
	"testing"
	"time"

	"github.com/advancevillage/3rd/logx"
	"github.com/advancevillage/3rd/mathx"
)

func Test_grpc(t *testing.T) {
	ctx := context.WithValue(context.TODO(), logx.TraceId, mathx.UUID())

	logger, err := logx.NewLogger("debug")
	if err != nil {
		t.Fatal(err)
		return
	}

	var data = map[string]struct {
		host string
		port int
	}{
		"case1": {
			host: "127.0.0.1",
			port: 1995,
		},
	}

	for n, v := range data {
		f := func(t *testing.T) {
			sctx, cancel := context.WithDeadline(ctx, time.Now().Add(time.Minute))
			go waitQuitSignal(cancel)

			s, err := NewGrpcServer(ctx, logger,
				WithAddr(v.host, v.port),
				WithGrpcService(NewHealthorRegister(ctx, logger)),
			)
			if err != nil {
				t.Fatal(err)
				return
			}

			s.Start()
			<-sctx.Done()
		}
		t.Run(n, f)
	}
}
