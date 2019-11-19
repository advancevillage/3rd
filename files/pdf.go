//author: richard
package files

import (
	"bytes"
	"github.com/SebastiaanKlippert/go-wkhtmltopdf"
)

func (pdf *PDFFile) CreateFileFromBuffer(filename string, html []byte) error {
	//需要预先配置 wkhtmltopdf 对应环境的版本环境
	//https://wkhtmltopdf.org/index.html
	engine, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		return err
	}
	//设置PDF选项
	engine.Dpi.Set(600)
	engine.NoCollate.Set(false)
	engine.PageSize.Set(wkhtmltopdf.PageSizeA4)
	//生成PDF
	engine.AddPage(wkhtmltopdf.NewPageReader(bytes.NewBuffer(html)))
	err = engine.Create()
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
	engine.Dpi.Set(600)
	engine.NoCollate.Set(false)
	engine.PageSize.Set(wkhtmltopdf.PageSizeA4)
	//生成PDF
	engine.AddPage(wkhtmltopdf.NewPage(url))
	err = engine.Create()
	if err != nil {
		return err
	}
	err = engine.WriteFile(filename)
	if err != nil {
		return err
	}
	return nil
}

