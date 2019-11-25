//author: richard
package pools

import (
	"3rd/logs"
	"sync"
)

const (
	PoolError = "pool is full"
)

type Handler func(e interface{})

type Pool interface {
	Process(e interface{}, f Handler) error
}

type GoPool struct {
	count  int32 		//池容量
	size   int32    	//运行数
	lock   sync.Mutex
	cond   *sync.Cond
	logger logs.Logs
}

type worker struct {
	g *GoPool
	f Handler
}