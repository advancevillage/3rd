//author: richard
package test

import (
	"github.com/advancevillage/3rd/https"
	"log"
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
		err := ctx.WriteCookie("cookie", "richard", "/", "localhost")
		if err != nil {
			t.Error(err.Error())
		}
		value, err := ctx.ReadCookie("cookie")
		body.Value = value
		if err != nil {
			t.Error(err.Error())
		}
		ctx.JsonResponse(http.StatusOK, body)
	}
	routers := []https.Router{
		{"GET", "/v1/test/ping", f},
	}

	p := func(ctx *https.Context) {
		headers := map[string]string {
			"Access-Control-Allow-Origin": "*",
		}
		ctx.Header(headers)
		ctx.Next()
	}
	server := https.NewServer("0.0.0.0", 13148, routers, p)
	err := server.StartServer()
	if err != nil {
		t.Error(err.Error())
	}
}

func TestServer_AwsLambda(t *testing.T) {
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
		err := ctx.WriteCookie("cookie", "richard", "/", "localhost")
		if err != nil {
			log.Println(err.Error())
		}
		if err != nil {
			log.Println(err.Error())
		}
		ctx.JsonResponse(http.StatusOK, body)
	}
	routers := []https.Router{
		{"GET", "/v1/test/ping", f},
	}
	server := https.NewAwsApiGatewayLambdaServer(routers)
	err := server.StartServer()
	if err != nil {
		log.Println(err.Error())
	}
}

func TestClient(t *testing.T) {
	client := https.NewRequest(nil, 60, 2)
	uploadFile := "111/test.pdf"
	extraParams := map[string]string{
		"title":       "My Document",
		"author":      "richard",
		"description": "A document with all the Go programming language secrets",
	}
	body, err := client.Upload("https://google.com/upload", nil, extraParams, uploadFile)
	if err != nil {
		t.Error(err.Error())
	} else {
		t.Log(string(body))
	}
}
