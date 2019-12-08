//author: richard
package test

import (
	"github.com/advancevillage/3rd/logs"
	"github.com/advancevillage/3rd/utils"
	"strconv"
	"testing"
)

func TestTxtLogger_Error(t *testing.T) {
	logger, err := logs.NewTxtLogger("error.log", 16, 2)
	if err != nil {
		t.Error(err.Error())
	}
	for i := 0; i < 100; i++ {
		str := strconv.Itoa(i) + "ddddddddddddddd"
		logger.Emergency(str)
	}
}

func BenchmarkTxtLogger(b *testing.B) {
	logger, err := logs.NewTxtLogger("info.log", 64, 2)
	if err != nil {
		b.Error(err.Error())
	}
	for i := 0; i < b.N; i++ {
		str := utils.SnowFlakeId()
		logger.Info("%d", str)
	}
}
