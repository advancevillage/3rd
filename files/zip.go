//author: richard
package files

import (
	"archive/zip"
	"errors"
	"io"
	"os"
	"path/filepath"
)

//@param: filename
//@eg: xxx/xxx.txt  or xxx/xxx.pdf
//output: xxx/xxx.zip
func (z *ZipFile) CreateFileFromBuffer(filename string, content []byte) error {
	err := CreatePath(filename)
	if err != nil {
		return err
	}
	dir := filepath.Dir(filename)
	base := filepath.Base(filename)
	ext := filepath.Ext(base)
	cf := dir + string(filepath.Separator) + base[:len(base)-len(ext)] + ZipExt
	c, err := os.Create(cf)
	if err != nil {
		return err
	}
	w := zip.NewWriter(c)
	f, err := w.Create(base)
	if err != nil {
		return err
	}
	_, err = f.Write(content)
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	return nil
}

func (z *ZipFile) CreateFileFromFile(filename string, src string) error {
	err := CreatePath(filename)
	if err != nil {
		return err
	}
	dir := filepath.Dir(filename)
	base := filepath.Base(filename)
	ext := filepath.Ext(base)
	cf := dir + string(filepath.Separator) + base[:len(base)-len(ext)] + ZipExt
	c, err := os.Create(cf)
	if err != nil {
		return err
	}
	w := zip.NewWriter(c)
	f, err := w.Create(base)
	if err != nil {
		return err
	}
	r, err := os.Open(src)
	if err != nil {
		return err
	}
	_, err = io.Copy(f, r)
	if err != nil {
		return err
	}
	e1 := w.Close()
	e2 := r.Close()
	if e1 != nil || e2 != nil {
		return errors.New("files close failed")
	}
	return nil
}
