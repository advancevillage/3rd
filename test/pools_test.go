//author: richard
package test

import (
	"3rd/logs"
	"3rd/pools"
	"3rd/utils"
	"encoding/json"
	"os"
	"testing"
)

func TestGoPool_Process(t *testing.T) {
	logger, err := logs.NewTxtLogger("go.log", 16, 4)
	if err != nil {
		t.Error(err.Error())
	}
	handler := func(e interface{}) {
		buf, err := json.Marshal(e)
		if err != nil {
			logger.Error(err.Error())
		}
		logger.Info(string(buf))
	}
	pool := pools.NewGoPool(4, logger)
	for i := 0; i < 1000; i ++ {
		err = pool.Process("richard", handler)
	}
}

func BenchmarkGoPool_Process(b *testing.B) {
	logger, err := logs.NewTxtLogger("go.log", 32, 4)
	if err != nil {
		b.Error(err.Error())
	}
	handler := func(e interface{}) {
		buf, err := json.Marshal(e)
		if err != nil {
			logger.Error(err.Error())
		}
		logger.Info(string(buf))
	}
	pool := pools.NewGoPool(10, logger)
	for i := 0; i < b.N; i++ {
		for i := 0; i < 10000; i ++ {
			err = pool.Process(utils.UUID(), handler)
		}
	}
}

func TestExit_Process(t *testing.T) {
	for {
		go func() { os.Exit(0) }()
	}
}