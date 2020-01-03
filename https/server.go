//author: richard
package https

import (
	"encoding/base64"
	"fmt"
	"github.com/advancevillage/3rd/files"
	"github.com/advancevillage/3rd/utils"
	"github.com/gin-gonic/gin"
	"io"
	"io/ioutil"
	"os"
)

func NewServer(host string, port int, router []Router, middleware ...Handler) *Server {
	s := Server{}
	s.host = host
	s.port = port
	s.router = router
	s.middleware = middleware
	//setting release mode
	gin.SetMode(gin.ReleaseMode)
	s.engine = gin.New()
	return &s
}

func (s *Server) StartServer() error {
	//init router
	for i := 0; i < len(s.router); i++ {
		s.handle(s.router[i].Method, s.router[i].Path, s.router[i].Func)
	}
	//init middleware
	handlers := make([]gin.HandlerFunc, 0, len(s.middleware))
	for i := range s.middleware {
		handler := func(ctx *gin.Context) {
			c := Context{ctx:ctx}
			s.middleware[i](&c)
		}
		handlers = append(handlers, handler)
	}
	s.engine.Use(handlers[:] ...)
	//run
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

func (c *Context) Body() ([]byte, error) {
	buf, err := ioutil.ReadAll(c.ctx.Request.Body)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

//@brief: application/json
//@param: code 状态码
//@param: body
func (c *Context) JsonResponse(code int, body interface{}) {
	c.ctx.JSON(code, body)
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

//@brief: 保存上传的文件 multipart/form-data
//@param:
func (c *Context) Save(filename string) error {
	_, fh, err := c.ctx.Request.FormFile("file")
	if err != nil {
		return err
	}
	in, err := fh.Open()
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()
	err = files.CreatePath(filename)
	if err != nil {
		return err
	}
	out, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()
	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return nil
}
