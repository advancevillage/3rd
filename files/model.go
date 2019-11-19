//author: richard
package files

type Files interface {
	CreateFileFromBuffer(filename string, content []byte) error
}

type XMLFile struct {}

type PDFFile struct {
	FontSize float64
}





