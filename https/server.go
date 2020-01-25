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
	//init middleware
	s.plugin(s.middleware)
	//init router
	for i := 0; i < len(s.router); i++ {
		s.handle(s.router[i].Method, s.router[i].Path, s.router[i].Func)
	}
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

func (s *Server) plugin(middleware []Handler) {
	handlers := make([]gin.HandlerFunc, 0, len(middleware))
	for i := range middleware {
		handler := func(ctx *gin.Context) {
			c := Context{ctx:ctx}
			middleware[i](&c)
		}
		handlers = append(handlers, handler)
	}
	s.engine.Use(handlers[:] ...)
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
func (c *Context) JSON(code int, body interface{}) {
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
func (c *Context) Save(filename string) (string, error) {
	_, fh, err := c.ctx.Request.FormFile("file")
	i := len(fh.Filename) - 1
	j := 0
	for ; i > 0; i-- {
		if fh.Filename[i] == '.' {
			break
		} else {
			continue
		}
	}
	for ; j < len(filename); j++ {
		if filename[j] == '.' {
			break
		} else {
			continue
		}
	}
	filename = filename[:j] + fh.Filename[i:]
	if err != nil {
		return filename, err
	}
	in, err := fh.Open()
	if err != nil {
		return filename, err
	}
	defer func() { _ = in.Close() }()
	err = files.CreatePath(filename)
	if err != nil {
		return filename, err
	}
	out, err := os.Create(filename)
	if err != nil {
		return filename, err
	}
	defer func() { _ = out.Close() }()
	_, err = io.Copy(out, in)
	if err != nil {
		return filename, err
	}
	return filename, nil
}

func (c *Context) Next() {
	c.ctx.Next()
}

func (c *Context) Abort() {
	c.ctx.Abort()
}

func (c *Context) WriteHeader(headers map[string]string) {
	for key := range headers {
		c.ctx.Header(key, headers[key])
	}
}

func (c *Context) ReadHeader(h string) string {
	return c.ctx.GetHeader(h)
}
