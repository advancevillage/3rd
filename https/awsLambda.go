//author: richard
package https

import (
	"bytes"
	"context"
	"encoding/base64"
	"github.com/advancevillage/3rd/logs"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

const defaultStatusCode = -1
const contentTypeHeaderKey = "Content-Type"

var (
	errorLambdaParseRequest = "lambda parse request error"
    errorHeaders = map[string]string {"Content-Type": "text/plain; charset=utf-8", "Connection": "close",}
)

var lambdaError = events.APIGatewayProxyResponse{
	StatusCode: http.StatusBadGateway,
	Headers: errorHeaders,
	Body: errorLambdaParseRequest,
	IsBase64Encoded: false,
}

func NewAwsApiGatewayLambdaServer(router []Router) *AwsApiGatewayLambdaServer {
	s := AwsApiGatewayLambdaServer{}
	s.logger = logs.NewStdLogger()
	s.router = router
	s.engine = gin.New()
	return &s
}

func (s *AwsApiGatewayLambdaServer) StartServer() error {
	//setting release mode
	gin.SetMode(gin.ReleaseMode)
	//init router
	for i := 0; i < len(s.router); i++ {
		s.handle(s.router[i].Method, s.router[i].Path, s.router[i].Func)
	}
	lambda.Start(s.lambdaHandler)
	return nil
}

func (s *AwsApiGatewayLambdaServer) handle(method string, path string, f Handler) {
	handler := func(ctx *gin.Context) {
		c := Context{ctx:ctx}
		f(&c)
	}
	s.engine.Handle(method, path, handler)
}

func (s *AwsApiGatewayLambdaServer) lambdaHandler(ctx context.Context, r events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	//ApiGateway服务入口
	var err error
	var response = events.APIGatewayProxyResponse{}
	body := []byte(r.Body)
	path := r.Path
	method := r.HTTPMethod
	if r.IsBase64Encoded {
		body, err = base64.StdEncoding.DecodeString(r.Body)
		if err != nil {
			s.logger.Error(err.Error())
			return	lambdaError, nil
		}
	}
	request, err := http.NewRequest(strings.ToUpper(method), path, bytes.NewReader(body))
	if err != nil {
		s.logger.Error(err.Error())
		return lambdaError, nil
	}
	request.Header = r.MultiValueHeaders
	//进入jin框架处理
	rw := &responseWriter{
		headers: make(http.Header),
		status:  defaultStatusCode,
	}
	s.engine.ServeHTTP(rw, request)
	//返回ApiGateway服务
	if rw.status == defaultStatusCode {
		return lambdaError, nil
	}
	response.IsBase64Encoded = false
	response.StatusCode = rw.status
	response.Body = string(rw.body.Bytes())
	response.MultiValueHeaders = rw.headers
	return response, nil
}

type responseWriter struct {
	headers http.Header
	body    bytes.Buffer
	status  int
}

func (rw *responseWriter) Header() http.Header {
	return rw.headers
}

func (rw *responseWriter) Write(body []byte) (int, error) {
	if rw.status == -1 {
		rw.status = http.StatusOK
	}
	if rw.Header().Get(contentTypeHeaderKey) == "" {
		rw.Header().Add(contentTypeHeaderKey, http.DetectContentType(body))
	}
	return (&rw.body).Write(body)
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
}