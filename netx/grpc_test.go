package netx

import (
	"context"
	"testing"
	"time"

	"github.com/advancevillage/3rd/logx"
	"github.com/advancevillage/3rd/mathx"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
	status "google.golang.org/grpc/status"
)

func Test_grpc(t *testing.T) {
	ctx := context.WithValue(context.TODO(), logx.TraceId, mathx.UUID())
	logger, err := logx.NewLogger("debug")
	assert.Nil(t, err)

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
			sctx, cancel := context.WithDeadline(ctx, time.Now().Add(time.Second*3))
			go waitQuitSignal(cancel)

			s, err := NewGrpcServer(sctx, logger,
				WithServerAddr(v.host, v.port),
				WithGrpcService(NewHealthorRegister(sctx, logger)),
			)
			assert.Nil(t, err)
			go s.Start()

			time.Sleep(time.Second)
			c, err := NewGrpcClient(sctx, logger,
				WithClientAddr(v.host, v.port),
			)
			assert.Nil(t, err)
			defer c.Close(sctx)

			healthor := NewHealthorClient(c.Conn(sctx))

			md := metadata.New(map[string]string{
				logx.TraceId: mathx.UUID(),
			})

			loggerCtx := metadata.NewOutgoingContext(sctx, md)

			stream, err := healthor.BidiPing(loggerCtx)
			if err != nil {
				st, ok := status.FromError(err)
				if ok {
					t.Logf("gRPC Status Code: %v\n", st.Code())
					t.Logf("gRPC Status Message: %v\n", st.Message())
				}
			}
			assert.Nil(t, err)

			go func() {
				for i := 0; i < 5; i++ {
					now := time.Now().UnixNano() / 1e6
					err = stream.Send(&PingRequest{T: now})
					assert.Nil(t, err)
				}
				err = stream.CloseSend()
				assert.Nil(t, err)
			}()

			for {
				reply, err := stream.Recv()
				if err != nil {
					break
				}
				t.Logf("reply.T: %d", reply.T)
			}

			assert.Nil(t, err)
			<-sctx.Done()
			t.Log(sctx.Err())
		}
		t.Run(n, f)
	}
}
