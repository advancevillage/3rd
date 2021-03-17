package netx

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"os"
	"os/signal"
	"syscall"
)

var (
	errConnectClosed = errors.New("connection closed")
	errConnected     = errors.New("connection connected")
	errConnecting    = errors.New("connection connecting")
	errReconnected   = errors.New("reconnection")
)

//@overview: 监听信号处理
func waitQuitSignal(cancel context.CancelFunc) {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c
	cancel()
}

type Handler func(context.Context, []byte) []byte

type ServerOption struct {
	TransportOption
	PriKey *ecdsa.PrivateKey //本地服务私钥
}

type ClientOption struct {
	TransportOption
	EnodeUrl string
	PriKey   *ecdsa.PrivateKey //本地服务私钥
}
