//author: richard
package files

import (
	"io/ioutil"
	"os"
)

func NewXMLFile() *XMLFile {
	return &XMLFile{}
}

func (x *XMLFile) CreateFileFromBuffer(filename string, content []byte) error {
	err := CreatePath(filename)
	if err != nil {
		return err
	}
	out, err := os.Create(filename)
	if err != nil {
		return err
	}
	_, err = out.Write(content)
	if err != nil {
		return err
	}
	err = out.Close()
	if err != nil {
		return err
	}
	return nil
}

func (x *XMLFile) ReadFile(filename string) ([]byte, error) {
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

