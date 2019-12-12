//author: richard
package test

import (
	"github.com/advancevillage/3rd/logs"
	"github.com/advancevillage/3rd/times"
	"github.com/advancevillage/3rd/utils"
	"sync"
	"testing"
	"time"
)

func TestRandsString(t *testing.T) {
	for i := 0; i < 5000; i++ {
		str := utils.RandsString(50)
		t.Log(str)
	}
}

func TestUUID(t *testing.T) {
	for i := 0; i < 100; i++ {
		str := utils.UUID()
		t.Log(str)
	}
}

func TestRandInt(t *testing.T) {
	for i := 0; i < 100; i++ {
		t.Log(utils.RandsInt(100))
	}
}

func TestTime(t *testing.T) {
	logger, err := logs.NewTxtLogger("storage.log", 512, 4)
	if err != nil {
		t.Error(err.Error())
	}
	for i := 0; i < 500; i++ {
		t.Log(times.NewTimes(logger).Timestamp())
		t.Log(times.NewTimes(logger).TimeString())
	}
}

func BenchmarkTime(b *testing.B) {
	logger, err := logs.NewTxtLogger("storage.log", 512, 4)
	if err != nil {
		b.Error(err.Error())
	}
	for i := 0; i < b.N; i++ {
		b.Log(times.NewTimes(logger).Timestamp())
		b.Log(times.NewTimes(logger).TimeString())
	}
}

func BenchmarkUUID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		str := utils.UUID()
		b.Log(str)
	}
}

func TestFunc(t *testing.T) {
	ff := func() func() int {
		var x int
		return func() int {
			x++
			return x * x
		}
	}
	f := ff()
	t.Log(f())
	t.Log(f())
	t.Log(f())
	t.Log(f())

}

func TestSnowFlake(t *testing.T) {
	for i := 0; i < 50; i++ {
		snowFlakeId := utils.SnowFlakeId()
		t.Log(snowFlakeId)
	}
}

func BenchmarkSnowFlake(b *testing.B) {
	for i := 0; i < b.N; i++ {
		snowFlakeId := utils.SnowFlakeId()
		b.Log(snowFlakeId)
	}
}

func TestSyncPool(t *testing.T) {
	pool := &sync.Pool{
		// 默认的返回值设置，不写这个参数，默认是nil
		New: func() interface{} {
			return 0
		},
	}
	// 看一下初始的值，这里是返回0，如果不设置New函数，默认返回nil
	init := pool.Get()
	t.Log(init)
	// 设置一个参数1
	pool.Put(1)
	// 获取查看结果
	num := pool.Get()
	t.Log(num)
	// 再次获取，会发现，已经是空的了，只能返回默认的值。
	num = pool.Get()
	t.Log(num)
}

func TestTimeFormat(t *testing.T) {
	logger, err := logs.NewTxtLogger("storage.log", 512, 4)
	if err != nil {
		t.Error(err.Error())
	}
	t.Log(times.NewTimes(logger).TimeFormatString(time.ANSIC))
	t.Log(times.NewTimes(logger).TimeFormatString(time.RFC1123Z))
	t.Log(times.NewTimes(logger).TimeFormatString(time.RFC1123))
}
