package dbx_test

import (
	"context"
	"testing"
	"time"

	"github.com/advancevillage/3rd/dbx"
	"github.com/advancevillage/3rd/logx"
	"github.com/advancevillage/3rd/mathx"
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
