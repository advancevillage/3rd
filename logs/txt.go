//author: richard
package logs

import (
	"bufio"
	"fmt"
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
	//文件名称最大字符长度
	length := len(m + level) + len("2019/11/27 13:48:10") + 25
	logger := log.New(txt.cache[txt.ptr % txt.total], level, log.LstdFlags | log.Lshortfile)
	for {
		free := txt.cache[txt.ptr % txt.total].Available()
		if length < free {
			logger.Println(m)
			break
		} else if length < txt.size {
			err = txt.cache[txt.ptr % txt.total].Flush()
			if err != nil {
				logger.SetPrefix(LogLevelEmer)
				logger.SetOutput(txt.file)
				logger.Println(err.Error())
			}
			txt.ptr = (txt.ptr + 1) % txt.total
			logger.SetOutput(txt.cache[txt.ptr % txt.total])
		} else {
			logger.SetOutput(txt.file)
			logger.Println(m)
			break
		}
	}
	return
}

//@brief: error log
func (txt *TxtLogger) Error(format string, a ...interface{}) {
	txt.Write(LogLevelEror, fmt.Sprintf(format, a...))
}

//@brief: warning log
func (txt *TxtLogger) Warning(format string, a ...interface{}) {
	txt.Write(LogLevelWarn, fmt.Sprintf(format, a...))
}

//@brief: debug log
func (txt *TxtLogger) Debug(format string, a ...interface{}) {
	txt.Write(LogLevelDebg, fmt.Sprintf(format, a...))
}

func (txt *TxtLogger) Info(format string, a ...interface{}) {
	txt.Write(LogLevelInfo, fmt.Sprintf(format, a...))
}

func (txt *TxtLogger) Alert(format string, a ...interface{}) {
	txt.Write(LogLevelAlrt, fmt.Sprintf(format, a...))
}

func (txt *TxtLogger) Critical(format string, a ...interface{}) {
	txt.Write(LogLevelCrit, fmt.Sprintf(format, a...))
}

func (txt *TxtLogger) Emergency(format string, a ...interface{}) {
	txt.Write(LogLevelEmer, fmt.Sprintf(format, a...))
}