//author: richard
package test

import (
	"3rd/logs"
	"testing"
)

func TestTxtLogger_Error(t *testing.T) {
	logger, err := logs.NewTxtLogger("error.log", 256, 2)
	if err != nil {
		t.Error(err.Error())
	}
	for i := 0; i < 100; i++ {
		str := "ddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd"
		logger.Error(str)
	}
}
