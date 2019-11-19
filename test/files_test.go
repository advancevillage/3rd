//author: richard
package test

import (
	"3rd/files"
	"bytes"
	"html/template"
	"testing"
)

func TestPDFFile_CreateFile(t *testing.T) {
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
	err = pdf.CreateFileFromUrl("test.pdf", "http://localhost:63342/3rd/test/test.html?_ijt=97n1893s3jb0ll4q4tcjm1gse5")
	if err != nil {
		t.Error(err.Error())
	}
	return
}


