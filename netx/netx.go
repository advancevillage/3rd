package netx

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin/binding"
)

const (
	X_Request_Latency = "X-Request-Latency"
	X_Request_Arrival = "X-Request-Arrival"
)

func waitQuitSignal(cancel context.CancelFunc) {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c
	cancel()
}

func ShouldBind(r *http.Request, obj any) error {
	return binding.JSON.Bind(r, obj)
}

func ShoudBindQuery(r *http.Request, obj any) error {
	return binding.Query.Bind(r, obj)
}
