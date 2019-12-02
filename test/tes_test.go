//author: richard
package test

import (
	"3rd/logs"
	"3rd/storages"
	"testing"
)

func TestNewTES(t *testing.T) {
	logger, err := logs.NewTxtLogger("tes.log", 128, 3)
	if err != nil {
		t.Error(err.Error())
		return
	}
	urls := []string{"http://localhost:9200"}
	tes, err := storages.NewTES(urls, logger)
	if err != nil {
		t.Error(err.Error())
		return
	}
	t.Log(tes)
}
