//author: richard
package logs

import (
	"fmt"
	"log"
	"os"
)

func NewStdLogger() *StdLogger {
	std := &StdLogger{}
	std.logger = log.New(os.Stdout, "", log.LstdFlags | log.Lshortfile)
	return std
}

func (std *StdLogger) Write(level string, m string) {
	std.logger.SetPrefix(level)
	std.logger.Println(m)
	return
}

//@brief: error log
func (std *StdLogger) Error(format string, a ...interface{}) {
	std.Write(LogLevelEror, fmt.Sprintf(format, a...))
}

//@brief: warning log
func (std *StdLogger) Warning(format string, a ...interface{}) {
	std.Write(LogLevelWarn, fmt.Sprintf(format, a...))
}

//@brief: debug log
func (std *StdLogger) Debug(format string, a ...interface{}) {
	std.Write(LogLevelDebg, fmt.Sprintf(format, a...))
}

func (std *StdLogger) Info(format string, a ...interface{}) {
	std.Write(LogLevelInfo, fmt.Sprintf(format, a...))
}

func (std *StdLogger) Alert(format string, a ...interface{}) {
	std.Write(LogLevelAlrt, fmt.Sprintf(format, a...))
}

func (std *StdLogger) Critical(format string, a ...interface{}) {
	std.Write(LogLevelCrit, fmt.Sprintf(format, a...))
}

func (std *StdLogger) Emergency(format string, a ...interface{}) {
	std.Write(LogLevelEmer, fmt.Sprintf(format, a...))
}
