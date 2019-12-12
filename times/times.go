//author: richard
package times

import (
	"github.com/advancevillage/3rd/logs"
	"time"
)

const (
	Layout = "2006-01-02 15:04:05"
)

type Times struct {
	logger logs.Logs
}

func NewTimes(logger logs.Logs) *Times {
	return &Times{logger:logger}
}

func (t *Times) Timestamp() int64 {
	return time.Now().Unix()
}

func (t *Times) TimeString() string {
	return time.Now().Format(Layout)
}

func (t *Times) TimeFormatString(layout string) string {
	return time.Now().Format(layout)
}

func (t *Times) FormatStringToTime(layout string, value string) (timestamp int64) {
	times, err := time.ParseInLocation(layout, value, time.Local)
	if err != nil {
		t.logger.Error(err.Error())
		return
	}
	timestamp = times.Unix()
	return
}



