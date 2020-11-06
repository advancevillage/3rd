//author: richard
package https

import (
	"github.com/advancevillage/3rd/logs"
	"github.com/gin-gonic/gin"
)

//@overview: 定义接口
type IServer interface {
	StartServer() error
}

type Handler func(*Context)

type Context struct {
	ctx *gin.Context
}

type Client struct {
	headers    map[string]string
	timeout    int64
	retryCount uint
}

type Router struct {
	Method string
	Path   string
	Func   Handler
}

type Server struct {
	host       string
	port       int
	router     []Router
	middleware []Handler
	engine     *gin.Engine
}

type AwsApiGatewayLambdaServer struct {
	engine     *gin.Engine
	logger     logs.Logs
	router     []Router
	middleware []Handler
}
