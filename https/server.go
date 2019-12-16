//author: richard
package https

import (
	"encoding/base64"
	"fmt"
	"github.com/advancevillage/3rd/utils"
	"github.com/gin-gonic/gin"
	"io/ioutil"
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

func (c *Context) Body() ([]byte, error) {
	buf, err := ioutil.ReadAll(c.ctx.Request.Body)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func (c *Context) WriteCookie(name string, value string, path string, domain string) error {
	maxAge := 2 * 3600 //秒
	cipherText, err := utils.EncryptUseAes([]byte(value))
	if err != nil {
		return err
	}
	text := base64.StdEncoding.EncodeToString(cipherText)
	c.ctx.SetCookie(name, text, maxAge, path, domain, false, true)
	return nil
}

func (c *Context) ReadCookie(name string) (string, error) {
	value, err := c.ctx.Cookie(name)
	if err != nil {
		return "", err
	}
	cipherText, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return "", err
	}
	plainText, err := utils.DecryptUseAes(cipherText)
	return string(plainText), err
}
