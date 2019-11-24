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
	f1 := func(r []byte, code byte) ([]byte, error) {
		log.Println(string(r), code, "f1")
		return []byte("f1"), nil
	}
	f2 := func(r []byte, code byte) ([]byte, error) {
		log.Println(string(r), code, )
		return []byte("f2"), nil
	}
	router := []wss.Router{
		wss.Router{Path:"/v1/f1/", Func: f1},
		wss.Router{Path:"/v1/f2/", Func:f2},
	}
	ws := wss.NewServer("0.0.0.0", 13147, router, logger)
	err = ws.StartServer()
	if err != nil {
		t.Error(err.Error())
	}
}
