//author: richard
package logs

import (
	"bufio"
	"log"
	"os"
	"sync"
)

const (
	Byte = 1
	KB	 = Byte * 1024
	MB   = KB * 1024

	LogLevelEmer  = "[emr]"		//系统级紧急
	LogLevelAlrt  = "[alt]"		//系统级警告
	LogLevelCrit  = "[crt]"		//系统级危险
	LogLevelEror  = "[err]"		//用户级错误
	LogLevelWarn  = "[wan]"		//用户级警告
	LogLevelInfo  = "[inf]"		//用户级重要
	LogLevelDebg  = "[dbg]"		//用户级调试
)

type Logs interface {
	Error(format string, a ...interface{})
	Warning(format string, a ...interface{})
	Debug(format string, a ...interface{})
	Info(format string, a ...interface{})
	Alert(format string, a ...interface{})
	Critical(format string, a ...interface{})
	Emergency(format string, a ...interface{})
	Write(level string, m string)
}

type TxtLogger struct {
	ptr 	   int
	size 	   int
	total 	   int
	filename   string
	cache      []*bufio.Writer
	file 	   *os.File
	mutex      sync.Mutex
	logger     *log.Logger
}

type StdLogger struct {
	logger    *log.Logger
}