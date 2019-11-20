//author: richard
package files

import "os"

func (xml *XMLFile) CreateFileFromBuffer(filename string, content []byte) error {
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

