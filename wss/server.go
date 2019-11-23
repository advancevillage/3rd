//author: richard
package wss

import (
	"3rd/logs"
	"fmt"
	"github.com/gobwas/ws"
	"io"
	"net"
)

func NewServer(host string, port int, logger logs.Logs, handler Handler) *Server {
	s := &Server{
		host:host,
		port: port,
		logger:logger,
		handler: handler,
	}
	return s
}

func (s *Server) StartServer() (err error) {
	ln, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.host, s.port))
	if err != nil {
		s.logger.Emergency(err.Error())
		return
	}
	for {
		s.conn, err = ln.Accept()
		if err != nil {
			s.logger.Emergency(err.Error())
			continue
		}
		s.hs, err = ws.Upgrade(s.conn)
		if err != nil {
			s.logger.Emergency(err.Error())
			continue
		}
		go s.server()
	}
}

func (s *Server) server() {
	defer func () { _ = s.conn.Close() }()
	for {
		//接受请求
		header, err := ws.ReadHeader(s.conn)
		if err != nil {
			s.logger.Error(err.Error())
		}
		payload := make([]byte, header.Length)
		_, err = io.ReadFull(s.conn, payload)
		if err != nil {
			s.logger.Error(err.Error())
		}
		if header.Masked {
			ws.Cipher(payload, header.Mask, 0)
		}
		header.Masked = false
		//处理请求
		err = s.handler(payload, byte(header.OpCode))
		if err != nil {
			s.logger.Error(err.Error())
		}
		//响应请求
		if err := ws.WriteHeader(s.conn, header); err != nil {
			s.logger.Error(err.Error())
		}
		if _, err := s.conn.Write(payload); err != nil {
			s.logger.Error(err.Error())
		}
		if header.OpCode == ws.OpClose {
			break
		}
	}
}