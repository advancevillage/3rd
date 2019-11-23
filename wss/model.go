//author: richard
package wss

import (
	"3rd/logs"
	"github.com/gobwas/ws"
	"net"
)

const (
	OpContinuation = ws.OpContinuation
	OpText         = ws.OpText
	OpBinary       = ws.OpBinary
	OpClose        = ws.OpClose
	OpPing         = ws.OpPing
	OpPong         = ws.OpPong
)


type OpCode ws.OpCode

type Handler func([]byte, byte) error

type Server struct {
	host 	string
	port 	int
	conn 	net.Conn
	logger  logs.Logs
	handler Handler
	hs 		ws.Handshake
}