package dbx_test

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/advancevillage/3rd/dbx"
	"github.com/advancevillage/3rd/logx"
	"github.com/advancevillage/3rd/mathx"
	"github.com/advancevillage/3rd/x"
	"github.com/stretchr/testify/assert"
)

func Test_Locker(t *testing.T) {
	logger, err := logx.NewLogger("debug")
	if err != nil {
		t.Fatal(err)
		return
	}

	var data = map[string]struct {
		key string
		val string
		ttl int64
	}{
		"case1": {
			key: mathx.RandStr(10),
			val: mathx.RandStr(10),
			ttl: 2,
		},
	}

	for n, v := range data {
		f := func(t *testing.T) {
			ctx := context.WithValue(context.TODO(), logx.TraceId, mathx.UUID())
			ctx, cancel := context.WithDeadline(ctx, time.Now().Add(time.Second*time.Duration(v.ttl<<1)))
			defer cancel()

			locker, err := dbx.NewCacheRedisLocker(ctx, logger)
			if err != nil {
				t.Fatal(err)
				return
			}
			// 加锁
			ok, err := locker.Lock(ctx, v.key, v.val, v.ttl)
			assert.Nil(t, err)
			assert.Equal(t, true, ok)
			// 非加锁客户端解锁
			ok, err = locker.Unlock(ctx, v.key, mathx.RandStr(10))
			assert.Nil(t, err)
			assert.Equal(t, false, ok)
			// 加锁客户端解锁
			ok, err = locker.Unlock(ctx, v.key, v.val)
			assert.Nil(t, err)
			assert.Equal(t, true, ok)

			// 加锁
			ok, err = locker.Lock(ctx, v.key, v.val, v.ttl)
			assert.Nil(t, err)
			assert.Equal(t, true, ok)
			time.Sleep(time.Second * time.Duration(v.ttl+1))
			// 超时解锁
			ok, err = locker.Unlock(ctx, v.key, v.val)
			assert.Nil(t, err)
			assert.Equal(t, false, ok)
		}
		t.Run(n, f)
	}

}

func Test_hash(t *testing.T) {
	logger, err := logx.NewLogger("debug")
	assert.Nil(t, err)
	ctx := context.TODO()
	var data = map[string]struct {
		key    string
		exp    time.Duration
		field  []x.Option
		incr   []x.Option
		decr   []x.Option
		expect []x.Option
	}{
		"case1": {
			key: mathx.RandStr(10),
			exp: time.Second * 2,
			field: []x.Option{
				x.WithKV("inf", "JDZEYdHmXAoRXMwUrDFyBrFFtvEyuUqO"),
			},
			incr: []x.Option{
				x.WithKV("cnt", 10),
				x.WithKV("sin", 123456),
			},
			decr: []x.Option{
				x.WithKV("cnt", -5),
				x.WithKV("sin", 4),
			},
			expect: []x.Option{
				x.WithKV("inf", "JDZEYdHmXAoRXMwUrDFyBrFFtvEyuUqO"),
				x.WithKV("cnt", "5"),
				x.WithKV("sin", "123460"),
			},
		},
	}
	rc, err := dbx.NewCacheRedis(ctx, logger)
	assert.Nil(t, err)

	for n, v := range data {
		f := func(t *testing.T) {
			h := rc.CreateHashCacher(ctx, v.key, v.exp)
			// 新增
			err = h.Set(ctx, x.NewBuilder(append(v.field, v.incr...)...))
			assert.Nil(t, err)

			// 自增
			err = h.Incr(ctx, x.NewBuilder(v.decr...))
			assert.Nil(t, err)

			// 获取
			b, err := h.Get(ctx, "inf", "cnt", "sin")
			assert.Nil(t, err)
			assert.Equal(t, x.NewBuilder(v.expect...).Build(), b.Build())

			// 删除
			err = h.Del(ctx, "sin")
			assert.Nil(t, err)

			// 获取
			b, err = h.Get(ctx, "inf", "cnt", "sin")
			assert.Nil(t, err)
			assert.NotEqual(t, x.NewBuilder(v.expect...).Build(), b.Build())

			// 过期
			time.Sleep(v.exp)

			// 获取
			b, err = h.Get(ctx, "inf", "cnt", "sin")
			assert.Nil(t, err)
			assert.Equal(t, x.NewBuilder().Build(), b.Build())
		}
		t.Run(n, f)
	}
}

func Test_string(t *testing.T) {
	logger, err := logx.NewLogger("debug")
	assert.Nil(t, err)
	ctx := context.TODO()
	var data = map[string]struct {
		key string
		exp time.Duration
		val any
	}{
		"case1": {
			key: mathx.RandStr(10),
			exp: time.Second,
			val: mathx.RandStrNum(50),
		},
	}
	rc, err := dbx.NewCacheRedis(ctx, logger)
	assert.Nil(t, err)

	for n, v := range data {
		f := func(t *testing.T) {
			str := rc.CreateStringCacher(ctx, v.key, v.exp)
			// 新增
			err = str.Set(ctx, v.val)
			assert.Nil(t, err)

			// 获取
			val, err := str.Get(ctx)
			assert.Nil(t, err)
			assert.Equal(t, v.val, val)

			// 删除
			err = str.Del(ctx)
			assert.Nil(t, err)

			// 获取
			val, err = str.Get(ctx)
			assert.Nil(t, err)
			assert.Equal(t, "", val)

			// 新增
			err = str.Set(ctx, v.val)
			assert.Nil(t, err)

			// 过期
			time.Sleep(v.exp + time.Millisecond)

			// 获取
			val, err = str.Get(ctx)
			assert.Nil(t, err)
			assert.Equal(t, "", val)
		}
		t.Run(n, f)
	}
}

func Test_string_incr(t *testing.T) {
	logger, err := logx.NewLogger("debug")
	assert.Nil(t, err)
	ctx := context.TODO()
	var data = map[string]struct {
		key  string
		exp  time.Duration
		val  any
		incr int64
	}{
		"case1": {
			key:  mathx.RandStr(10),
			exp:  time.Millisecond * 500,
			val:  int64(99),
			incr: 2,
		},
	}
	rc, err := dbx.NewCacheRedis(ctx, logger)
	assert.Nil(t, err)

	for n, v := range data {
		f := func(t *testing.T) {
			str := rc.CreateStringCacher(ctx, v.key, v.exp)
			// 新增
			err = str.Set(ctx, v.val)
			assert.Nil(t, err)

			// 自增
			err = str.Incr(ctx, v.incr)
			assert.Nil(t, err)

			// 获取
			val, err := str.Get(ctx)
			assert.Nil(t, err)
			act, err := strconv.ParseInt(val, 10, 64)
			assert.Nil(t, err)
			assert.Equal(t, v.val.(int64)+v.incr, act)

			// 删除
			err = str.Del(ctx)
			assert.Nil(t, err)

			// 获取
			val, err = str.Get(ctx)
			assert.Nil(t, err)
			assert.Equal(t, "", val)

			// 新增
			err = str.Set(ctx, v.val)
			assert.Nil(t, err)

			// 过期
			time.Sleep(v.exp + time.Millisecond)

			// 获取
			val, err = str.Get(ctx)
			assert.Nil(t, err)
			assert.Equal(t, "", val)
		}
		t.Run(n, f)
	}
}
