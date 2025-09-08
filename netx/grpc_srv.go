package netx

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/advancevillage/3rd/logx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

var _ Server = (*grpcSrv)(nil)

type grpcSrv struct {
	opts   serverOptions
	logger logx.ILogger

	srv     *grpc.Server       // grpc server
	rctx    context.Context    // root context
	rcancel context.CancelFunc // root cancel
}

func newGrpcSrv(ctx context.Context, logger logx.ILogger, opt ...ServerOption) (*grpcSrv, error) {
	// 0. 设置配置
	opts := defaultServerOptions
	for _, o := range opt {
		o.apply(&opts)
	}

	// 1. 服务对象
	s := &grpcSrv{logger: logger, opts: opts}

	// 2. 上下文
	s.rctx, s.rcancel = context.WithCancel(ctx)

	// 3. 安全凭证
	creds, err := credentials.NewServerTLSFromFile(s.opts.crt, s.opts.key)
	if err != nil {
		s.logger.Errorw(s.rctx, "read cert file failed", "err", err, "crt", s.opts.crt, "key", s.opts.key)
		return nil, err
	}

	// 4. 构建服务
	kaep := keepalive.EnforcementPolicy{
		MinTime:             5 * time.Second, // 客户端 ping 的最小间隔
		PermitWithoutStream: true,            // 允许空闲时心跳
	}

	kasp := keepalive.ServerParameters{
		MaxConnectionIdle:     2 * time.Minute,  // 空闲多久断开
		MaxConnectionAge:      5 * time.Minute,  // 连接存活最长时间
		MaxConnectionAgeGrace: 1 * time.Minute,  // 超时后宽限期
		Time:                  15 * time.Second, // 发送 ping 的间隔
		Timeout:               5 * time.Second,  // ping 等待响应超时时间
	}

	s.srv = grpc.NewServer(
		grpc.Creds(creds),
		grpc.KeepaliveParams(kasp),
		grpc.KeepaliveEnforcementPolicy(kaep),
	)

	// 5. 注册服务
	for i := range s.opts.ss {
		s.opts.ss[i](s.srv)
	}

	return s, nil
}

func (s *grpcSrv) Start() {
	go s.start()
	go waitQuitSignal(s.rcancel)
	<-s.rctx.Done()
	s.logger.Infow(s.rctx, "grpc server closed", "host", s.opts.host, "port", s.opts.port)
	s.srv.GracefulStop()
	time.Sleep(time.Second)
}

func (s *grpcSrv) start() {
	// 1. 监听端口
	var listener, err = net.Listen("tcp", fmt.Sprintf("%s:%d", s.opts.host, s.opts.port))
	if err != nil {
		s.logger.Errorw(s.rctx, "listen failed", "err", err, "host", s.opts.host, "port", s.opts.port)
		return
	}
	s.logger.Infow(s.rctx, "grpc server start", "host", s.opts.host, "port", s.opts.port)
	err = s.srv.Serve(listener)
	if err != nil {
		s.logger.Errorw(s.rctx, "serve failed", "err", err, "host", s.opts.host, "port", s.opts.port)
		return
	}
}
