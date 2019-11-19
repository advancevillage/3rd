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
	txt.total = total
	txt.cache = make([]*bufio.Writer, 0, total)
	txt.filename = filename
	txt.file, err = os.Create(filename)
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
		free := txt.cache[txt.ptr % txt.total].Available() - 2
		if free > mlen {
			logger.Println(m)
			break
		} else {
			logger.Println(m[:free])
			err = txt.cache[txt.ptr % txt.total].Flush()
			if err != nil {
				m = err.Error()
				level  = LogLevelEmer
				logger = log.New(txt.file, level, log.LstdFlags | log.Lshortfile)
				logger.Println(m)
				break
			}
			txt.ptr = (txt.ptr + 1) % txt.total
			m = m[free:]
		}
		logger.SetOutput(txt.cache[txt.ptr % txt.total])
		mlen -= free
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
