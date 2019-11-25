//author: richard
package rpcs

import (
	"3rd/logs"
	"fmt"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
)

func NewServer(host string, port int, logger logs.Logs, rcvr []interface{}) *Server {
	s := &Server{
		host:host,
		port: port,
		logger:logger,
		rcvr:rcvr,
	}
	return s
}

//@link: https://www.cnblogs.com/hangxin1940/p/3256995.html
func (s *Server) StartServer() (err error) {
	server := rpc.NewServer()
	ln, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.host, s.port))
	if err != nil {
		s.logger.Error(err.Error())
		return err
	}
	for i := 0; i < len(s.rcvr); i++ {
		err = server.Register(s.rcvr[i])
		if err != nil {
			s.logger.Error(err.Error())
		}
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			s.logger.Error(err.Error())
			continue
		}
		go server.ServeCodec(jsonrpc.NewServerCodec(conn))
	}
}