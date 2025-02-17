package dbx

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/advancevillage/3rd/logx"
	"github.com/advancevillage/3rd/x"
	"github.com/redis/go-redis/v9"
)

type CacheLocker interface {
	Lock(ctx context.Context, key string, val string, ttl int64) (bool, error)
	Unlock(ctx context.Context, key string, val string) (bool, error)
}

type CacheOption interface {
	apply(*cacheOption)
}

func WithCacheDNS(dsn string) CacheOption {
	return newFuncCacheOption(func(o *cacheOption) {
		o.dsn = dsn
	})
}

func WithCacheThreshold(threshold int) CacheOption {
	return newFuncCacheOption(func(o *cacheOption) {
		o.threshold = threshold
	})
}

type cacheOption struct {
	dsn       string // 组件地址
	threshold int    // 阈值
}

var defaultCacheOptions = cacheOption{
	dsn:       "redis://127.0.0.1:6379/0?protocol=3",
	threshold: 10,
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

type redisClient struct {
	opts   cacheOption
	rdb    *redis.Client
	logger logx.ILogger
}

func newRedisClient(ctx context.Context, logger logx.ILogger, opt ...CacheOption) (*redisClient, error) {
	// 1. 解析参数
	opts := defaultCacheOptions
	for _, o := range opt {
		o.apply(&opts)
	}
	// 2. 创建连接
	rdbOpts, err := redis.ParseURL(opts.dsn)
	if err != nil {
		logger.Errorw(ctx, "redis parse url failed", "err", err, "dsn", opts.dsn)
		return nil, err
	}
	rdb := redis.NewClient(rdbOpts)
	// 3. 验证连接
	err = rdb.Ping(ctx).Err()
	if err != nil {
		logger.Errorw(ctx, "redis ping failed", "err", err, "dsn", opts.dsn)
		return nil, err
	}
	logger.Infow(ctx, "cache redis connect success", "dsn", opts.dsn)

	// 4. 返回对象
	return &redisClient{
		opts:   opts,
		rdb:    rdb,
		logger: logger,
	}, nil
}

var _ CacheLocker = (*redisLocker)(nil)

type redisLocker struct {
	redisClient
}

func NewCacheRedisLocker(ctx context.Context, logger logx.ILogger, opt ...CacheOption) (CacheLocker, error) {
	return newCacheRedisLocker(ctx, logger, opt...)
}

func newCacheRedisLocker(ctx context.Context, logger logx.ILogger, opt ...CacheOption) (*redisLocker, error) {
	rc, err := newRedisClient(ctx, logger, opt...)
	if err != nil {
		return nil, err
	}
	return &redisLocker{*rc}, nil
}

func (c *redisLocker) Lock(ctx context.Context, key string, val string, ttl int64) (bool, error) {
	ok, err := c.rdb.SetNX(ctx, key, val, time.Duration(ttl)*time.Second).Result()
	if err != nil {
		c.logger.Errorw(ctx, "redis lock failed", "err", err, "key", key, "val", val)
		return false, err
	}
	return ok, nil
}

// Unlock 释放锁，只有持有锁的客户端才能释放
// 解锁不存在或已过期的锁会返回false
func (c *redisLocker) Unlock(ctx context.Context, key string, val string) (bool, error) {
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

type Cacher interface {
	CreateHashCacher(ctx context.Context, key string, exp time.Duration) HashCacher
}

const (
	cmd_CACHE_HINCR = "hincr"
	cmd_CACHE_HSET  = "hset"
)

var _ Cacher = (*redisCacher)(nil)

type redisCacher struct {
	redisClient
}

func NewCacheRedis(ctx context.Context, logger logx.ILogger, opt ...CacheOption) (Cacher, error) {
	return newCacheRedis(ctx, logger, opt...)
}

func newCacheRedis(ctx context.Context, logger logx.ILogger, opt ...CacheOption) (*redisCacher, error) {
	rc, err := newRedisClient(ctx, logger, opt...)
	if err != nil {
		return nil, err
	}
	return &redisCacher{*rc}, nil
}

func (c *redisCacher) CreateHashCacher(ctx context.Context, key string, exp time.Duration) HashCacher {
	return newHashRedisCacher(ctx, c, key, exp)
}

func (c *redisCacher) hPipe(ctx context.Context, cmd string, key string, kv map[string]interface{}) error {
	var (
		pipe = c.rdb.TxPipeline()
		succ int
		fail int
		st   = time.Now().UnixNano() / 1e6
	)
	for k, v := range kv {
		switch cmd {
		case cmd_CACHE_HINCR:
			pipe.HIncrBy(ctx, key, k, v.(int64))

		case cmd_CACHE_HSET:
			pipe.HSet(ctx, key, k, v)
		}
	}

	var (
		r, err = pipe.Exec(ctx)
		et     = time.Now().UnixNano() / 1e6
	)
	if err != nil {
		c.logger.Errorw(ctx, "redis pipe hset failed", "err", err, "key", key, "kvCnt", len(kv))
		return err
	}
	for i := range r {
		if r[i].Err() != nil {
			fail++
		} else {
			succ++
		}
	}
	c.logger.Infow(ctx, "redis pipe hset stat", "key", key, "succCnt", succ, "failCnt", fail, "delay", et-st)
	return nil
}

type HashCacher interface {
	Get(ctx context.Context, fields ...string) (x.Builder, error)
	Del(ctx context.Context, fields ...string) error
	Set(ctx context.Context, b x.Builder) error
	Incr(ctx context.Context, b x.Builder) error
}

var _ HashCacher = (*hashRedisCacher)(nil)

type hashRedisCacher struct {
	rc  *redisCacher
	key string
	exp time.Duration
}

func newHashRedisCacher(ctx context.Context, rc *redisCacher, key string, exp time.Duration) HashCacher {
	return &hashRedisCacher{
		rc:  rc,
		key: key,
		exp: exp,
	}
}

func (c *hashRedisCacher) Get(ctx context.Context, fields ...string) (x.Builder, error) {
	// 1. 参数检查
	if len(fields) <= 0 {
		return x.NewBuilder(), nil
	}
	// 2. 执行
	r, err := c.rc.rdb.HMGet(ctx, c.key, fields...).Result()
	if err != nil {
		c.rc.logger.Errorw(ctx, "redis hash get failed", "err", err, "key", c.key, "fileds", fields)
		return nil, err
	}
	kv := []x.Option{}
	for i, field := range fields {
		if r[i] == nil {
			c.rc.logger.Warnw(ctx, "redis hash get empty", "key", c.key, "field", field)
			continue
		}
		kv = append(kv, x.WithKV(field, r[i]))
	}
	return x.NewBuilder(kv...), nil
}

func (c *hashRedisCacher) Set(ctx context.Context, b x.Builder) error {
	var kv = b.Build()
	if len(kv) <= 0 {
		return nil
	}
	err := c.rc.hPipe(ctx, cmd_CACHE_HSET, c.key, kv)
	if err != nil {
		return err
	}
	if c.exp <= 0 {
		return nil
	}
	err = c.rc.rdb.Expire(ctx, c.key, c.exp).Err()
	if err != nil {
		c.rc.logger.Errorw(ctx, "redis expire failed", "err", err, "key", c.key)
		return err
	}
	return nil
}

func (c *hashRedisCacher) Del(ctx context.Context, fields ...string) error {
	if len(fields) <= 0 {
		return nil
	}
	r, err := c.rc.rdb.HDel(ctx, c.key, fields...).Result()
	if err != nil {
		c.rc.logger.Errorw(ctx, "redis hash del failed", "err", err, "key", c.key, "fields", fields)
		return err
	}
	c.rc.logger.Infow(ctx, "redis hash del success", "key", c.key, "fields", fields, "cnt", r)
	return nil
}

func (c *hashRedisCacher) Incr(ctx context.Context, b x.Builder) error {
	var (
		kv = b.Build()
		ki = make(map[string]interface{}, len(kv))
	)
	for k, v := range kv {
		i, err := strconv.ParseInt(fmt.Sprint(v), 10, 64)
		if err != nil {
			c.rc.logger.Warnw(ctx, "parse int64 failed", "err", err, "key", c.key, "field", k, "value", v)
			return err
		}
		ki[k] = i
	}
	if len(ki) <= 0 {
		return nil
	}
	return c.rc.hPipe(ctx, cmd_CACHE_HINCR, c.key, ki)
}
