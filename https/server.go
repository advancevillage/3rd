//author: richard
package https

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

func NewServer(host string, port int, router []Router) *Server {
	s := Server{}
	s.host = host
	s.port = port
	s.router = router
	s.engine = gin.New()
	return &s
}

func (s *Server) StartServer() error {
	//setting release mode
	gin.SetMode(gin.ReleaseMode)
	//init router
	for i := 0; i < len(s.router); i++ {
		s.handle(s.router[i].Method, s.router[i].Path, s.router[i].Func)
	}
	err := s.engine.Run(fmt.Sprintf("%s:%d", s.host, s.port))
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) handle(method string, path string, f Handler) {
	handler := func(ctx *gin.Context) {
		c := Context{ctx:ctx}
		f(&c)
	}
	s.engine.Handle(method, path, handler)
}

//@param q 查询参数
//level: postform > query > path > ctx.Set()
func (c *Context) Param(q string) string {
	value := c.ctx.PostForm(q)
	if len(value) == 0 {
		value = c.ctx.Query(q)
	}
	if len(value) == 0 {
		value = c.ctx.Param(q)
	}
	if len(value) == 0 {
		value = c.ctx.GetString(q)
	}
	return value
}

//@brief: application/json
//@param: code 状态码
//@param: body
func (c *Context) JsonResponse(code int, body interface{}) {
	c.ctx.JSON(code, body)
}