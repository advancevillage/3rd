package netx

import (
	"context"
	"io"
	"time"

	"github.com/advancevillage/3rd/logx"
	"github.com/advancevillage/3rd/mathx"
	grpc "google.golang.org/grpc"
)

type GrpcRegister func(grpc.ServiceRegistrar)

type Server interface {
	Start()
}

func NewGrpcServer(ctx context.Context, logger logx.ILogger, opt ...ServerOption) (Server, error) {
	return newGrpcSrv(ctx, logger, opt...)
}

type ServerOption interface {
	apply(*serverOptions)
}

func WithAddr(host string, port int) ServerOption {
	return newFuncServerOption(func(o *serverOptions) {
		o.host = host
		o.port = port
	})
}

func WithCredential(crt, key string) ServerOption {
	return newFuncServerOption(func(o *serverOptions) {
		o.crt = crt
		o.key = key
	})
}

func WithGrpcService(f GrpcRegister) ServerOption {
	return newFuncServerOption(func(o *serverOptions) {
		o.ss = append(o.ss, f)
	})
}

type serverOptions struct {
	ss   []GrpcRegister // 注册gRPC服务
	host string         // 服务地址
	port int            // 端口
	crt  string         // 证书 文件
	key  string         // 私钥 文件
}

var defaultServerOptions = serverOptions{
	host: "127.0.0.1",
	port: 13147,
	crt:  "cert.pem",
	key:  "privkey.pem",
	ss:   make([]GrpcRegister, 0, 1),
}

type funcServerOption struct {
	f func(*serverOptions)
}

func (fdo *funcServerOption) apply(do *serverOptions) {
	fdo.f(do)
}

func newFuncServerOption(f func(*serverOptions)) *funcServerOption {
	return &funcServerOption{
		f: f,
	}
}

var _ HealthorServer = (*healthorService)(nil)

type healthorService struct {
	logger logx.ILogger
	UnimplementedHealthorServer
}

func NewHealthorRegister(ctx context.Context, logger logx.ILogger) GrpcRegister {
	s := &healthorService{logger: logger}
	s.logger.Infow(ctx, "grpc healthor service register success", "now", time.Now().Unix())
	return s.Register
}

func (s *healthorService) Ping(ctx context.Context, req *PingRequest) (*PingReply, error) {
	return &PingReply{T: time.Now().UnixNano() / 1e6}, nil
}

func (s *healthorService) CPing(stream Healthor_CPingServer) error {
	var (
		ctx   = context.WithValue(stream.Context(), logx.TraceId, mathx.UUID())
		count = 0
	)
	for {
		var (
			req, err = stream.Recv()
			now      = time.Now().UnixNano() / 1e6
		)
		count += 1
		if err == io.EOF {
			s.logger.Infow(ctx, "stream finished", "count", count, "send", now)
			return stream.SendAndClose(&PingReply{T: now})
		}
		if err != nil {
			s.logger.Errorw(ctx, "stream quit quit unexpectedly", "err", err, "count", count, "receive", req.T)
			return err
		}
	}
}

func (s *healthorService) SPing(req *PingRequest, stream Healthor_SPingServer) error {
	var (
		ctx   = context.WithValue(stream.Context(), logx.TraceId, mathx.UUID())
		count = 2
	)
	for i := 0; i < count; i++ {
		var now = time.Now().UnixNano() / 1e6
		err := stream.Send(&PingReply{T: now})
		if err != nil {
			s.logger.Errorw(ctx, "stream quit unexpectedly", "err", err, "count", count, "send", now)
			return err
		}
		time.Sleep(time.Second)
	}
	s.logger.Infow(ctx, "stream finished", "count", count)
	return nil
}

func (s *healthorService) BidiPing(stream Healthor_BidiPingServer) error {
	var (
		ctx   = context.WithValue(stream.Context(), logx.TraceId, mathx.UUID())
		count = 0
	)

	for {
		var (
			req, err = stream.Recv()
			now      = time.Now().UnixNano() / 1e6
		)
		count += 1
		if err == io.EOF {
			s.logger.Infow(ctx, "stream finished", "count", count, "send", now)
			return nil
		}
		if err != nil {
			s.logger.Errorw(ctx, "stream quit unexpectedly", "err", err, "count", count, "receive", req.T)
			return err
		}
		err = stream.Send(&PingReply{T: now})
		if err != nil {
			s.logger.Errorw(ctx, "stream quit unexpectedly", "err", err, "count", count, "send", now)
			return err
		}
		s.logger.Infow(ctx, "stream receive", "count", count, "send", now)
	}
}

func (r *healthorService) Register(s grpc.ServiceRegistrar) {
	RegisterHealthorServer(s, r)
}
