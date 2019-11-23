//author: richard
package test

import (
	"3rd/logs"
	"3rd/wss"
	"log"
	"testing"
)

func TestWsServer_StartServer(t *testing.T) {
	logger, err := logs.NewTxtLogger("ws.log", 256, 4)
	if err != nil {
		t.Error(err.Error())
	}
	f := func(r []byte, code byte) error {
		log.Println(string(r), code)
		return nil
	}
	ws := wss.NewServer("0.0.0.0", 13147, logger, f)
	err = ws.StartServer()
	if err != nil {
		t.Error(err.Error())
	}
}
