//author: richard
package swagger

import (
	"bytes"
	"encoding/json"
	"html/template"
	"io/ioutil"
	"os"
)

func Parse(filename string) (*ISwagger, error) {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	swag := ISwagger{}
	err = json.Unmarshal(buf, &swag)
	if err != nil {
		return nil, err
	}
	return &swag, nil
}

func (s *ISwagger) ToHtml(tmp string, dest string) error {
	body, err := template.ParseFiles(tmp)
	if err != nil {
		return err
	}
	buf := bytes.Buffer{}
	err = body.Execute(&buf, s)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(dest, buf.Bytes(), os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}
