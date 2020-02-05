//author: richard
package test

import (
	"bytes"
	"fmt"
	"github.com/advancevillage/3rd/files"
	"html/template"
	"os"
	"os/exec"
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

func TestCommand(t *testing.T) {
	var stdoutBuf bytes.Buffer	//标准输出流
	var stderrBuf bytes.Buffer	//标准错误流
	var stdinBuf  bytes.Reader	//标准输入流
	cmd := exec.Command("python", "screen_record.py", "-t", "0000", "-f", "./222 333/test.mp4", "start")
	cmd.Stdin  = &stdinBuf
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf
	if err := cmd.Start(); err != nil {
		t.Error(err.Error())
		return
	}
	if err := cmd.Wait(); err != nil {
		t.Error(err.Error())
		return
	}
	t.Log(string(stdoutBuf.Bytes()), string(stderrBuf.Bytes()))
}