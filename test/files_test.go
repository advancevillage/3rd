//author: richard
package test

import (
	"3rd/files"
	"bytes"
	"fmt"
	"html/template"
	"testing"
)

func TestPDFFile_CreateFileFromBuffer(t *testing.T) {
	pdf := files.PDFFile{DPI: 800}
	body, err := template.ParseFiles("test.html")
	if err != nil {
		t.Error(err.Error())
	}
	html := new(bytes.Buffer)
	err = body.Execute(html, body)
	if err != nil {
		t.Error(err.Error())
	}
	err = pdf.CreateFileFromUrl("111/test.pdf", "https://medium.com/@mena.meseha/what-is-the-difference-between-mysql-innodb-b-tree-index-and-hash-index-ed8f2ce66d69")
	if err != nil {
		t.Error(err.Error())
	}
	return
}

func TestZipFile_CreateFileFromBuffer(t *testing.T) {
	zip := files.ZipFile{}
	body, err := template.ParseFiles("test.html")
	if err != nil {
		t.Error(err.Error())
	}
	html := new(bytes.Buffer)
	err = body.Execute(html, body)
	if err != nil {
		t.Error(err.Error())
	}
	err = zip.CreateFileFromBuffer("111/test.txt", html.Bytes())
	if err != nil {
		t.Error(err.Error())
	}
	return
}

func TestBase_01 (t *testing.T) {
	int_chan := make(chan int, 1)
	string_chan := make(chan string)
	int_chan <- 1
	string_chan <- "string"
	select {
	case value := <- int_chan:
		fmt.Println(value)
	case value := <- string_chan:
			fmt.Println(value)
	}
}

