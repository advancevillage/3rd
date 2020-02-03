//author: richard
package test

import (
	"bytes"
	"fmt"
	"github.com/advancevillage/3rd/files"
	"html/template"
	"os"
	"path/filepath"
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
	err = pdf.CreateFileFromUrl("111/test.pdf", "https://www.cnblogs.com/jin-xin/articles/10268923.html")
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

func TestZipFile_CreateFileFromFile(t *testing.T) {
	zip := files.ZipFile{}
	err := zip.CreateFileFromFile("111/111.mp4", "/Users/sun/Pictures/视频素材/test.mp4")
	if err != nil {
		t.Error(err.Error())
	}
}

func TestSpaceFilepath(t *testing.T) {
	path := "./111/ni hao /test.mp4"
	dir := filepath.Dir(path)
	t.Log(dir)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		t.Error(err.Error())
	}
}
