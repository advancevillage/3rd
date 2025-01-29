package dbx

import (
	"context"
	"time"

	"github.com/advancevillage/3rd/logx"
	"github.com/redis/go-redis/v9"
)

type CacheLocker interface {
	Lock(ctx context.Context, key string, val string, ttl int64) (bool, error)
	Unlock(ctx context.Context, key string, val string) (bool, error)
}

type CacheOption interface {
	apply(*cacheOption)
}

func WithCacheDNS(dns string) CacheOption {
	return newFuncCacheOption(func(o *cacheOption) {
		o.dns = dns
	})
}

type cacheOption struct {
	dns string // 组件地址
}

var defaultCacheOptions = cacheOption{
	dns: "redis://127.0.0.1:6379/0?protocol=3",
}

type funcCacheOption struct {
	f func(*cacheOption)
}

func (fdo *funcCacheOption) apply(do *cacheOption) {
	fdo.f(do)
}

func newFuncCacheOption(f func(*cacheOption)) *funcCacheOption {
	return &funcCacheOption{
		f: f,
	}
}

var _ CacheLocker = (*cacheRedis)(nil)

type cacheRedis struct {
	opts   cacheOption
	rdb    *redis.Client
	logger logx.ILogger
}

func NewCacheRedisLocker(ctx context.Context, logger logx.ILogger, opt ...CacheOption) (CacheLocker, error) {
	return newCacheRedis(ctx, logger, opt...)
}

func newCacheRedis(ctx context.Context, logger logx.ILogger, opt ...CacheOption) (*cacheRedis, error) {
	// 1. 解析参数
	opts := defaultCacheOptions
	for _, o := range opt {
		o.apply(&opts)
	}
	// 2. 创建连接
	rdbOpts, err := redis.ParseURL(opts.dns)
	if err != nil {
		logger.Errorw(ctx, "redis parse url failed", "err", err, "dns", opts.dns)
		return nil, err
	}
	rdb := redis.NewClient(rdbOpts)
	// 3. 验证连接
	err = rdb.Ping(ctx).Err()
	if err != nil {
		logger.Errorw(ctx, "redis ping failed", "err", err, "dns", opts.dns)
		return nil, err
	}
	logger.Infow(ctx, "cache redis connect success", "dns", opts.dns)

	// 4. 返回对象
	return &cacheRedis{
		opts:   opts,
		rdb:    rdb,
		logger: logger,
	}, nil
}

func (c *cacheRedis) Lock(ctx context.Context, key string, val string, ttl int64) (bool, error) {
	ok, err := c.rdb.SetNX(ctx, key, val, time.Duration(ttl)*time.Second).Result()
	if err != nil {
		c.logger.Errorw(ctx, "redis lock failed", "err", err, "key", key, "val", val)
		return false, err
	}
	return ok, nil
}

// Unlock 释放锁，只有持有锁的客户端才能释放
// 解锁不存在或已过期的锁会返回false
func (c *cacheRedis) Unlock(ctx context.Context, key string, val string) (bool, error) {
	lua := `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		else
			return 0
		end
	`
	result, err := c.rdb.Eval(ctx, lua, []string{key}, val).Result()
	if err != nil {
		return false, err
	}
	r := result.(int64)
	// key不存在或者不属于当前客户端
	if r != 1 {
		return false, nil
	}
	return true, nil
}
