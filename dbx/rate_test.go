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

// Test_rate_window: redis 固定窗口内放行 / 超限拒绝 / 窗口过期后恢复
func Test_rate_window(t *testing.T) {
	logger, err := logx.NewLogger("debug")
	assert.Nil(t, err)

	data := map[string]struct {
		prefix    string
		key       string
		threshold int64
		burst     int
	}{
		"case1": {
			prefix:    "rate:test:",
			key:       mathx.RandStr(10),
			threshold: 5,
			burst:     100, // 本地桶放宽, 让 redis 成为限流点
		},
	}

	for n, v := range data {
		f := func(t *testing.T) {
			ctx := context.WithValue(context.TODO(), logx.TraceId, mathx.UUID())

			l, err := dbx.NewRateLimiter(ctx, logger, dbx.WithRateRule(v.prefix, v.threshold, v.burst))
			assert.Nil(t, err)

			// 窗口内前 threshold 次放行
			for i := int64(0); i < v.threshold; i++ {
				ok, err := l.Allow(ctx, v.key)
				assert.Nil(t, err)
				assert.Equal(t, true, ok)
			}
			// 超阈值被拒
			ok, err := l.Allow(ctx, v.key)
			assert.Nil(t, err)
			assert.Equal(t, false, ok)

			// 等窗口(1s)过期后恢复
			time.Sleep(time.Second + time.Millisecond*200)
			ok, err = l.Allow(ctx, v.key)
			assert.Nil(t, err)
			assert.Equal(t, true, ok)
		}
		t.Run(n, f)
	}
}

// Test_rate_allowN: WithAllowN 单次消耗多个令牌
func Test_rate_allowN(t *testing.T) {
	logger, err := logx.NewLogger("debug")
	assert.Nil(t, err)

	data := map[string]struct {
		prefix    string
		key       string
		threshold int64
		burst     int
		n         int
	}{
		"case1": {
			prefix:    "rate:test:",
			key:       mathx.RandStr(10),
			threshold: 10,
			burst:     100,
			n:         10,
		},
	}

	for n, v := range data {
		f := func(t *testing.T) {
			ctx := context.WithValue(context.TODO(), logx.TraceId, mathx.UUID())

			l, err := dbx.NewRateLimiter(ctx, logger, dbx.WithRateRule(v.prefix, v.threshold, v.burst))
			assert.Nil(t, err)

			// 一次消耗 n 个令牌, 恰好用满阈值
			ok, err := l.Allow(ctx, v.key, dbx.WithAllowN(v.n))
			assert.Nil(t, err)
			assert.Equal(t, true, ok)

			// 再取 1 个即超限
			ok, err = l.Allow(ctx, v.key)
			assert.Nil(t, err)
			assert.Equal(t, false, ok)
		}
		t.Run(n, f)
	}
}

// Test_rate_wait: WithAllowWait 阻塞等待本地令牌补充后放行
func Test_rate_wait(t *testing.T) {
	logger, err := logx.NewLogger("debug")
	assert.Nil(t, err)

	data := map[string]struct {
		prefix    string
		key       string
		threshold int64
		burst     int
	}{
		"case1": {
			prefix:    "rate:test:",
			key:       mathx.RandStr(10),
			threshold: 2, // limit = 2 tokens/s => 每 ~500ms 补一个
			burst:     1, // 本地桶容量 1, 让阻塞可观测
		},
	}

	for n, v := range data {
		f := func(t *testing.T) {
			ctx := context.WithValue(context.TODO(), logx.TraceId, mathx.UUID())
			ctx, cancel := context.WithTimeout(ctx, time.Second*3)
			defer cancel()

			l, err := dbx.NewRateLimiter(ctx, logger, dbx.WithRateRule(v.prefix, v.threshold, v.burst))
			assert.Nil(t, err)

			// 先耗尽本地桶(burst=1)
			ok, err := l.Allow(ctx, v.key)
			assert.Nil(t, err)
			assert.Equal(t, true, ok)

			// 阻塞等待本地补充令牌(~500ms)后放行
			st := time.Now()
			ok, err = l.Allow(ctx, v.key, dbx.WithAllowWait())
			cost := time.Since(st)
			assert.Nil(t, err)
			assert.Equal(t, true, ok)
			assert.GreaterOrEqual(t, cost, time.Millisecond*300)
		}
		t.Run(n, f)
	}
}
