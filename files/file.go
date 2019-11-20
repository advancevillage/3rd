//author: richard
package files

import (
	"os"
	"path"
)

func CreatePath(filename string) error {
	dir := path.Dir(filename)
	_, err := os.Stat(dir)
	if err != nil {
		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return err
		} else {
			return nil
		}
	} else {
		return nil
	}
}
