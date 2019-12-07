//author: richard
package test

import (
	"github.com/advancevillage/3rd/https"
	"net/http"
	"testing"
)

func TestServer_StartServer(t *testing.T) {
	f := func(ctx *https.Context) {
		body := struct {
			Value string `json:"value"`
			AccountId string `json:"accountId"`
			Pong string `json:"pong"`
		}{
			Value: "richard@cuger.com",
			AccountId: "1aa79ea8-0c3f-11ea-9753-0242ac120002",
			Pong: "pong",
		}
		ctx.JsonResponse(http.StatusOK, body)
	}
	routers := []https.Router{
		{"GET", "/v1/test/ping", f},
	}
	server := https.NewServer("0.0.0.0", 13147, routers)
	err := server.StartServer()
	if err != nil {
		t.Error(err.Error())
	}
}
