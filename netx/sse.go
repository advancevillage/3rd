package netx

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type sseSrv struct {
}

// https://cloud.tencent.com/developer/article/2532395
func (s *sseSrv) write(ctx *gin.Context, r *http.Request) (HttpResponse, error) {
	// 1. 透传数据，上游处理
	data := ""

	ctx.SSEvent("message", "hello world")

	// 2. 返回协议数据
	h := http.Header{}
	h.Add("Content-Type", "text/event-stream")
	h.Add("Cache-Control", "no-cache")
	h.Add("Connection", "keep-alive")
	return newHttpResponse([]byte(data+"\n\n"), h, http.StatusOK), nil
}
