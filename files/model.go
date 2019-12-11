//author: richard
package files

//文件后缀
const (
	ZipExt = ".zip"
)

type Files interface {
	CreateFileFromBuffer(filename string, content []byte) error
	ReadFile(filename string) ([]byte, error)
}

type XMLFile struct {}

type PDFFile struct {
	DPI uint
}

type ZipFile struct {}







