//author: richard
package logs

import (
	"bufio"
	"os"
)

const (
	Byte = 1
	KB	 = Byte * 1024
	MB   = KB * 1024

	LogLevelEmer  = "[emer]"		//系统级紧急
	LogLevelAlrt  = "[alrt]"		//系统级警告
	LogLevelCrit  = "[crit]"		//系统级危险
	LogLevelEror  = "[eror]"		//用户级错误
	LogLevelWarn  = "[warn]"		//用户级警告
	LogLevelInfo  = "[info]"		//用户级重要
	LogLevelDebg  = "[debg]"		//用户级调试
)

type Logs interface {
	Error(message string)
	Warning(message string)
	Debug(message string)
	Info(message string)
	Alert(message string)
	Critical(message string)
	Emergency(message string)
	Write(level string, m string)
}

type TxtLogger struct {
	ptr 	   int
	size 	   int
	total 	   int
	filename   string
	cache      []*bufio.Writer
	file 	   *os.File
}
