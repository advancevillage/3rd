//author: richard
package pools

import (
	"errors"
	"github.com/advancevillage/3rd/logs"
	"sync"
	"sync/atomic"
)

func NewGoPool(count int32, logger logs.Logs) *GoPool {
	g := &GoPool{
		count:count,
		size: 0,
		logger:logger,
	}
	g.cond = sync.NewCond(&g.lock)
	return g
}

func (g *GoPool) Process(e interface{}, f Handler) (err error) {
	g.cond.L.Lock()
	if g.size < g.count {
		w := &worker{g:g, f:f}
		go w.run(e)
		g.cond.L.Unlock()
		return nil
	}else {
		g.cond.Wait()
		g.cond.L.Unlock()
		return errors.New(PoolError)
	}
}

func (w *worker) run(e interface{}) {
	atomic.AddInt32(&w.g.size, 1)
	w.f(e)
	atomic.AddInt32(&w.g.size, -1)
	if w.g.size < w.g.count {
		w.g.cond.Signal()
	}
}