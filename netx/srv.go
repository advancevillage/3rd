package netx

import (
	"context"
	"io"
	"strings"
	"time"

	"github.com/advancevillage/3rd/logx"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
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

func WithServerAddr(host string, port int) ServerOption {
	return newFuncServerOption(func(o *serverOptions) {
		o.host = host
		o.port = port
	})
}

func WithServerCredential(crt, key string) ServerOption {
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
	now := time.Now().UnixNano() / 1e6
	ctx = s.withLoggerContext(ctx)
	s.logger.Infow(ctx, "stream finished", "count", 1, "reply", now)
	return &PingReply{T: now}, nil
}

func (s *healthorService) CPing(stream Healthor_CPingServer) error {
	var (
		ctx   = s.withLoggerContext(stream.Context())
		count = 0
	)
	for {
		var (
			req, err = stream.Recv()
			now      = time.Now().UnixNano() / 1e6
		)
		if err == io.EOF {
			s.logger.Infow(ctx, "stream finished", "count", count, "reply", now)
			return stream.SendAndClose(&PingReply{T: now})
		}
		count += 1
		if err != nil {
			s.logger.Errorw(ctx, "stream quit unexpectedly", "err", err, "count", count)
			return err
		}
		s.logger.Infow(ctx, "streaming", "count", count, "receive", req.T)
	}
}

func (s *healthorService) SPing(req *PingRequest, stream Healthor_SPingServer) error {
	var (
		ctx   = s.withLoggerContext(stream.Context())
		count = 2
	)
	for i := 0; i < count; i++ {
		var now = time.Now().UnixNano() / 1e6
		err := stream.Send(&PingReply{T: now})
		if err != nil {
			s.logger.Errorw(ctx, "stream quit unexpectedly", "err", err, "count", i, "reply", now)
			return err
		}
		time.Sleep(time.Second)
	}
	s.logger.Infow(ctx, "stream finished", "count", count)
	return nil
}

func (s *healthorService) BidiPing(stream Healthor_BidiPingServer) error {
	var (
		ctx   = s.withLoggerContext(stream.Context())
		count = 0
	)

	for {
		var (
			req, err = stream.Recv()
			now      = time.Now().UnixNano() / 1e6
		)
		if err == io.EOF {
			s.logger.Infow(ctx, "stream finished", "count", count, "reply", now)
			return nil
		}
		count += 1
		if err != nil {
			s.logger.Errorw(ctx, "stream quit unexpectedly", "err", err, "count", count)
			return err
		}
		err = stream.Send(&PingReply{T: now})
		if err != nil {
			s.logger.Errorw(ctx, "stream quit unexpectedly", "err", err, "count", count, "reply", now)
			return err
		}
		s.logger.Infow(ctx, "streaming", "count", count, "reply", now, "receive", req.T)
	}
}

func (r *healthorService) Register(s grpc.ServiceRegistrar) {
	RegisterHealthorServer(s, r)
}

func (s *healthorService) withLoggerContext(ctx context.Context) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx
	}
	traceIds := md.Get(logx.TraceId)
	return context.WithValue(ctx, logx.TraceId, strings.Join(traceIds, ","))
}
