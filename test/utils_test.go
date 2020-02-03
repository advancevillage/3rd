//author: richard
package test

import (
	"fmt"
	"github.com/advancevillage/3rd/times"
	"github.com/advancevillage/3rd/utils"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestRandsString(t *testing.T) {
	for i := 0; i < 5; i++ {
		str := utils.RandsString(50)
		t.Log(str)
	}
	for i := 0; i < 5; i++ {
		str := utils.RandsNumberString(50)
		t.Log(str)
	}
}

func TestUUID(t *testing.T) {
	for i := 0; i < 10; i++ {
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
	for i := 0; i < 500; i++ {
		t.Log(times.Timestamp())
		t.Log(times.TimeString())
	}
}

func BenchmarkTime(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.Log(times.Timestamp())
		b.Log(times.TimeString())
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
	t.Log(times.TimeFormatString(time.ANSIC))
	t.Log(times.TimeFormatString(time.RFC1123Z))
	t.Log(times.TimeFormatString(time.RFC1123))
	t.Log(times.TimeFormatString(times.YYYYMMddHHmmss))
}

func TestAesEncryptAndDecrypt(t *testing.T) {
	plaintext := []byte("我 爱 你")
	t.Log(plaintext)
	cipher, err := utils.EncryptUseAes(plaintext)
	if err != nil {
		t.Error(err.Error())
	}
	t.Log(string(cipher))
	plaintext, err = utils.DecryptUseAes(cipher)
	if err != nil {
		t.Error(err.Error())
	}
	t.Log(plaintext)
}

func TestReflect(t *testing.T) {
	var key = make([]string, 2, 5)
	key[0] = "11"
	key[1] = "22"
	tt := reflect.TypeOf(key)
	t.Log(tt,fmt.Sprintf("%#v", key))
}

func BenchmarkAesEncryptAndDecrypt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		plaintext := []byte(utils.RandsString(12))
		cipher, err := utils.EncryptUseAes(plaintext)
		if err != nil {
			b.Error(err.Error())
		}
		plaintext, err = utils.DecryptUseAes(cipher)
		if err != nil {
			b.Error(err.Error())
		}
	}
}

