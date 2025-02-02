package netx

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

const (
	X_Request_Latency = "X-Request-Latency"
	X_Request_Arrival = "X-Request-Arrival"
)

// @overview: 监听信号处理
func waitQuitSignal(cancel context.CancelFunc) {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c
	cancel()
}
