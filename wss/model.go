//author: richard
package wss

import (
	"3rd/logs"
	"github.com/gobwas/ws"
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
type Handler func([]byte, byte) ([]byte, error)

type Router struct {
	Path 	string
	Func 	Handler
}

type Server struct {
	host 	string
	port 	int
	logger  logs.Logs
	router  []Router
}