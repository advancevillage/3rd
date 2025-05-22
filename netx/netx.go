package netx

import (
	"context"
	"mime/multipart"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin/binding"
)

const (
	X_Request_Latency = "X-3rd-Request-Latency"
	X_Request_Arrival = "X-3rd-Request-Arrival"
	X_Request_Server  = "X-3rd-Request-Server"
)

const (
	rEQUEXT_CTX = "Inner-Request-Ctx"
)

func waitQuitSignal(cancel context.CancelFunc) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c
	cancel()
}

func ShouldBind(r *http.Request, obj any) error {
	return binding.Default(r.Method, r.Header.Get("Content-Type")).Bind(r, obj)
}

func ShoudBindQuery(r *http.Request, obj any) error {
	return binding.Query.Bind(r, obj)
}

func ShouldBindForm(r *http.Request, obj any) error {
	return binding.Form.Bind(r, obj)
}

func ShoudBindFile(r *http.Request, name string) (*multipart.FileHeader, error) {
	if r.MultipartForm == nil {
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			return nil, err
		}
	}
	f, fh, err := r.FormFile(name)
	if err != nil {
		return nil, err
	}
	f.Close()
	return fh, err
}
