package dbx

import (
	"context"
	"sync"
	"time"

	"github.com/advancevillage/3rd/logx"
	"github.com/advancevillage/3rd/x"
	"golang.org/x/time/rate"
)

//@overview: 两级限流器
//	1. 本地令牌桶(golang.org/x/time/rate): 进程内, 无网络开销, 作为前置过滤/突发吸收, 先挡一道保护 redis
//	2. redis 固定窗口(INCR): 跨实例的权威计数, 作为分布式硬上限
//	请求需两级都放行才通过

//------------------------------ 构造 Option ------------------------------

type RateOption = x.Options[rateOption]

func WithRateDsn(dsn string) RateOption {
	return x.NewFuncOptions(func(o *rateOption) {
		o.dsn = dsn
	})
}

// WithRateRule 统一配置限流规则: key 前缀 / 每窗口(1s)全局上限 / 本地令牌桶容量
func WithRateRule(prefix string, threshold int64, burst int) RateOption {
	return x.NewFuncOptions(func(o *rateOption) {
		o.prefix = prefix
		o.threshold = threshold
		o.burst = burst
	})
}

type rateOption struct {
	dsn       string        // redis 组件地址
	prefix    string        // key 命名空间
	threshold int64         // 每 window 窗口内的全局上限
	burst     int           // 本地令牌桶容量
	window    time.Duration // 固定窗口时长, 内部固定 1s, 不对外暴露
}

var defaultRateOption = rateOption{
	dsn:       "redis://127.0.0.1:6379/0?protocol=3",
	prefix:    "rate:",
	threshold: 100,
	burst:     100,
	window:    time.Second,
}

//------------------------------ 调用 Option ------------------------------

type AllowOption = x.Options[allowOption]

// WithAllowN 单次请求消耗的令牌数, 默认 1
func WithAllowN(n int) AllowOption {
	return x.NewFuncOptions(func(o *allowOption) {
		o.n = max(n, 1)
	})
}

// WithAllowWait 阻塞等待本地令牌, 直到取到或 ctx 结束; 默认非阻塞
func WithAllowWait() AllowOption {
	return x.NewFuncOptions(func(o *allowOption) {
		o.wait = true
	})
}

type allowOption struct {
	n    int
	wait bool
}

var defaultAllowOption = allowOption{
	n:    1,
	wait: false,
}

//------------------------------ 限流器 ------------------------------

type Limiter interface {
	Allow(ctx context.Context, key string, opt ...AllowOption) (bool, error)
}

// redis 固定窗口 Lua: 原子 INCRBY, 仅首次(窗口内第一次计数)设置过期, 超阈值返回 0
const rate_LIMIT_LUA = `
local c = redis.call("INCRBY", KEYS[1], ARGV[2])
if c == tonumber(ARGV[2]) then
	redis.call("PEXPIRE", KEYS[1], ARGV[3])
end
if c > tonumber(ARGV[1]) then
	return 0
end
return 1
`

var _ Limiter = (*redisRateLimiter)(nil)

type redisRateLimiter struct {
	*redisClient // 复用 cache.go 的连接封装(parse/ping/日志)
	opts         rateOption
	locals       sync.Map // key -> *rate.Limiter, 懒加载; 注意: key 基数大时会持续增长
}

func NewRateLimiter(ctx context.Context, logger logx.ILogger, opt ...RateOption) (Limiter, error) {
	return newRateLimiter(ctx, logger, opt...)
}

func newRateLimiter(ctx context.Context, logger logx.ILogger, opt ...RateOption) (*redisRateLimiter, error) {
	// 1. 解析参数
	opts := defaultRateOption
	for _, o := range opt {
		o.Apply(&opts)
	}

	// 2. 复用 cache.go 的 redis 连接
	rc, err := newRedisClient(ctx, logger, WithCacheDsn(opts.dsn))
	if err != nil {
		return nil, err
	}
	return &redisRateLimiter{
		redisClient: rc,
		opts:        opts,
	}, nil
}

// local 懒加载 key 对应的本地令牌桶; window 固定 1s => limit 即 threshold tokens/s
func (r *redisRateLimiter) local(key string) *rate.Limiter {
	if v, ok := r.locals.Load(key); ok {
		return v.(*rate.Limiter)
	}
	limit := rate.Limit(float64(r.opts.threshold) / r.opts.window.Seconds())
	actual, _ := r.locals.LoadOrStore(key, rate.NewLimiter(limit, r.opts.burst))
	return actual.(*rate.Limiter)
}

func (r *redisRateLimiter) Allow(ctx context.Context, key string, opt ...AllowOption) (bool, error) {
	// 1. 解析参数
	opts := defaultAllowOption
	for _, o := range opt {
		o.Apply(&opts)
	}
	// 2. 本地令牌桶: 非阻塞取或阻塞等待
	l := r.local(key)
	if opts.wait {
		if err := l.WaitN(ctx, opts.n); err != nil {
			r.logger.Infow(ctx, "rate local wait failed", "key", key, "n", opts.n, "err", err)
			return false, err
		}
	} else if !l.AllowN(time.Now(), opts.n) {
		r.logger.Infow(ctx, "rate local reject", "key", key, "n", opts.n)
		return false, nil
	}
	// 3. redis 固定窗口: 跨实例硬上限, 非阻塞
	return r.redisAllow(ctx, key, opts.n)
}

func (r *redisRateLimiter) redisAllow(ctx context.Context, key string, n int) (bool, error) {
	var (
		rk  = r.opts.prefix + key
		win = r.opts.window.Milliseconds()
	)
	result, err := r.rdb.Eval(ctx, rate_LIMIT_LUA, []string{rk}, r.opts.threshold, n, win).Result()
	if err != nil {
		r.logger.Errorw(ctx, "rate redis eval failed", "err", err, "key", rk, "n", n)
		return false, err
	}
	allow := result.(int64) == 1
	if allow {
		r.logger.Infow(ctx, "rate allow", "key", rk, "n", n, "threshold", r.opts.threshold)
	} else {
		r.logger.Infow(ctx, "rate redis reject", "key", rk, "n", n, "threshold", r.opts.threshold)
	}
	return allow, nil
}
