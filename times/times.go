//author: richard
package times

import (
	"time"
)

const (
	Layout = "2006-01-02 15:04:05"
	YYYYMMddHHmmss = "20060102150405"
)

func Timestamp() int64 {
	return time.Now().Unix()
}

func TimeString() string {
	return time.Now().Format(Layout)
}

func TimeFormatString(layout string) string {
	return time.Now().Format(layout)
}



