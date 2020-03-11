//author: richard
package test

import (
	"github.com/advancevillage/3rd/swagger"
	"testing"
)

func TestParseSwaggerJson(t *testing.T) {
	swag, err := swagger.Parse("swagger.json")
	if err != nil {
		t.Error(err.Error())
	}
	err = swag.ToHtml("swagger.html", "api.html")
	if err != nil {
		t.Error(err.Error())
	}
}