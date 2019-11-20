//author: richard
package logs

import (
	"bufio"
	"log"
	"os"
)

func NewTxtLogger(filename string, size int, total int) (*TxtLogger, error) {
	var err error
	txt := &TxtLogger{}
	txt.ptr = 0
	txt.size = size
	txt.total = total
	txt.cache = make([]*bufio.Writer, 0, total)
	txt.filename = filename
	txt.file, err = os.OpenFile(filename, os.O_CREATE | os.O_APPEND | os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	for i := 0; i < total; i++ {
		writer := bufio.NewWriterSize(txt.file, size)
		txt.cache = append(txt.cache,  writer)
	}
	return txt, nil
}

func (txt *TxtLogger) Write(level string, m string) {
	var err error
	mlen := len(m + level)
	logger := log.New(txt.cache[txt.ptr % txt.total], level, log.LstdFlags | log.Lshortfile)
	for {
		free := txt.cache[txt.ptr % txt.total].Available() - 1
		if mlen < free {
			logger.Println(m)
			break
		} else if mlen < txt.size {
			err = txt.cache[txt.ptr % txt.total].Flush()
			if err != nil {
				logger.SetPrefix(LogLevelEmer)
				logger.SetOutput(txt.file)
				logger.Println(err.Error())
			}
			txt.ptr = (txt.ptr + 1) % txt.total
		} else {
			logger.SetOutput(txt.file)
			logger.Println(m)
			break
		}
	}
	return
}

//@brief: error log
func (txt *TxtLogger) Error(message string) {
	txt.Write(LogLevelEror, message)
}

//@brief: warning log
func (txt *TxtLogger) Warning(message string) {
	txt.Write(LogLevelWarn, message)
}

//@brief: debug log
func (txt *TxtLogger) Debug(message string) {
	txt.Write(LogLevelDebg, message)
}

func (txt *TxtLogger) Info(message string) {
	txt.Write(LogLevelInfo, message)
}

func (txt *TxtLogger) Alert(message string) {
	txt.Write(LogLevelAlrt, message)
}

func (txt *TxtLogger) Critical(message string) {
	txt.Write(LogLevelCrit, message)
}

func (txt *TxtLogger) Emergency(message string) {
	txt.Write(LogLevelEmer, message)
}
