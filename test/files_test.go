//author: richard
package test

import (
	"bytes"
	"fmt"
	"github.com/advancevillage/3rd/files"
	"github.com/advancevillage/3rd/utils"
	"github.com/mholt/archiver/v3"
	"html/template"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
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

func TestRarFileParse(t *testing.T) {
	rar := archiver.NewRar()
	filename := "资质图片.rar"
	uuid := utils.RandsString(4)
	//先解压压缩文件
	err := rar.Unarchive(filename, uuid)
	if err != nil {
		t.Error(err.Error())
	}
	defer func() { _ = rar.Close() }()
	//i/o 操作较慢
	time.Sleep(time.Second)
	//分析文件目录
	dir := uuid + string(filepath.Separator) +  filename[:len(filename)-4]
	folders, err := ioutil.ReadDir(dir)
	if err != nil {
		t.Error(err.Error())
		return
	}

	for i := range folders {
		folder := folders[i]
		name := folder.Name()
		if !folder.IsDir() {
			continue
		}
		path := dir + string(filepath.Separator) + name
		fs, err := ioutil.ReadDir(path)
		if err != nil {
			continue
		}
		//遍历文件夹下的文件
		for j := range fs {
			f := fs[j]
			if f.IsDir() {
				continue
			}
			filename := f.Name()
			t.Log(fmt.Sprintf("%s  %s", name, filename))
		}
	}
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