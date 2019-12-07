//author: richard
package wss

import (
	"fmt"
	"github.com/advancevillage/3rd/logs"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"net/http"
)

func NewServer(host string, port int, router []Router, logger logs.Logs, ) *Server {
	s := &Server{
		host:host,
		port: port,
		logger:logger,
		router: router,
	}
	return s
}

func (s *Server) StartServer() (err error) {
	//init router
	for i := 0; i < len(s.router); i++ {
		s.handle(s.router[i].Path, s.router[i].Func)
	}
	err = http.ListenAndServe(fmt.Sprintf("%s:%d", s.host, s.port), nil)
	if err != nil {
		s.logger.Error(err.Error())
	}
	return
}

func (s *Server) server(w http.ResponseWriter, r *http.Request, handler Handler) {
	conn, _, hs, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		s.logger.Error(err.Error())
		s.logger.Error(hs.Protocol)
		return
	}
	go func () {
		defer func() { err = conn.Close() }()
		for {
			//接受请求
			payload, op, err := wsutil.ReadClientData(conn)
			if err != nil {
				s.logger.Error(err.Error())
				break
			}
			//处理请求
			payload, err = handler(payload, byte(op))
			if err != nil {
				s.logger.Error(err.Error())
			}
			//响应请求
			err = wsutil.WriteServerMessage(conn, op, payload)
			if err != nil {
				s.logger.Error(err.Error())
				break
			}
		}
	}()
}

func (s *Server) handle(path string, handler Handler) {
	f := func(w http.ResponseWriter, r *http.Request) {
		s.server(w, r, handler)
	}
	http.HandleFunc(path, f)
}