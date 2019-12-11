//author: richard
package files

import (
	"bytes"
	"github.com/SebastiaanKlippert/go-wkhtmltopdf"
	"io/ioutil"
	"os"
)

func (pdf *PDFFile) CreateFileFromBuffer(filename string, html []byte) error {
	//需要预先配置 wkhtmltopdf 对应环境的版本环境
	//https://wkhtmltopdf.org/index.html
	engine, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		return err
	}
	//设置PDF选项
	engine.Dpi.Set(pdf.DPI)
	engine.PageSize.Set(wkhtmltopdf.PageSizeA4)
	//生成PDF
	engine.AddPage(wkhtmltopdf.NewPageReader(bytes.NewBuffer(html)))
	err = engine.Create()
	if err != nil {
		return err
	}
	err = CreatePath(filename)
	if err != nil {
		return err
	}
	err = engine.WriteFile(filename)
	if err != nil {
		return err
	}
	return nil
}

func (pdf *PDFFile) CreateFileFromUrl(filename string, url string) error {
	//需要预先配置 wkhtmltopdf 对应环境的版本环境
	//https://wkhtmltopdf.org/index.html
	engine, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		return err
	}
	//设置PDF选项
	engine.Dpi.Set(pdf.DPI)
	engine.PageSize.Set(wkhtmltopdf.PageSizeA4)
	//生成PDF
	engine.AddPage(wkhtmltopdf.NewPage(url))
	err = engine.Create()
	if err != nil {
		return err
	}
	err = CreatePath(filename)
	if err != nil {
		return err
	}
	err = engine.WriteFile(filename)
	if err != nil {
		return err
	}
	return nil
}

func (pdf *PDFFile) ReadFile(filename string) ([]byte, error) {
	fd, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	buf, err := ioutil.ReadAll(fd)
	if err != nil {
		_ = fd.Close()
		return nil, err
	}
	_ = fd.Close()
	return buf, nil
}

