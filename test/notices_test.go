//author: richard
package test

import (
	"3rd/notices"
	"bytes"
	"html/template"
	"testing"
)

func TestEmail_SendMailUsingTLS(t *testing.T) {
	email := notices.NewTlsEmail("smtp.xxxx.com", "notice@xxx.com", "xxx.")
	body, err := template.ParseFiles("test.html")
	if err != nil {
		t.Error(err.Error())
	}
	html := new(bytes.Buffer)
	err = body.Execute(html, body)
	if err != nil {
		t.Error(err.Error())
	}
	err = email.SendMailUsingTLS([]string{"xxx@xxx.com",}, "ShowU", "example", html.Bytes())
	if err != nil {
		t.Error(err.Error())
	}
}
