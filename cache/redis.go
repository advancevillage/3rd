//author: richard
package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/advancevillage/3rd/proto/cacheProxy"

	"github.com/go-redis/redis/v8"
)

type ICache interface {
	//hash
	HashSet(ctx context.Context, key string, fields map[string]interface{}, ttl int) error
	HashGet(ctx context.Context, key string, field string) ([]byte, error)
	HashGetAll(ctx context.Context, key string) (map[string]string, error)
	//pipeline hash
	PHashGetAll(ctx context.Context, key ...string) (map[string]string, error)
}

type redisClient struct {
	cli redis.Cmdable
}

func NewCacheClient(cfg *cacheProxy.CacheOpt) (ICache, error) {
	if cfg.Single != nil {
		return newSingleRedisClient(cfg), nil
	} else if cfg.FailOver != nil {
		return newFailoverClient(cfg), nil
	} else if cfg.Cluster != nil {
		return newClusterClient(cfg), nil
	}
	return nil, fmt.Errorf("single, failOver, cluster options should have one valid config")

}

func newSingleRedisClient(cfg *cacheProxy.CacheOpt) *redisClient {
	cli := redis.NewClient(&redis.Options{
		Addr: cfg.GetSingle().GetAddr(),

		Password:        cfg.GetPassword(),
		MaxRetries:      int(cfg.GetMaxRetries()),
		MinRetryBackoff: milliToDuration(cfg.GetMinRetryBackoff()),
		MaxRetryBackoff: milliToDuration(cfg.GetMaxRetryBackoff()),
		DialTimeout:     milliToDuration(cfg.GetDialTimeout()),
		ReadTimeout:     milliToDuration(cfg.GetReadTimeout()),
		WriteTimeout:    milliToDuration(cfg.GetWriteTimeout()),
		PoolSize:        int(cfg.GetPoolSize()),
		MinIdleConns:    int(cfg.GetMinIdleConns()),

		MaxConnAge:         milliToDuration(cfg.GetMaxConnAge()),
		PoolTimeout:        milliToDuration(cfg.GetPoolTimeout()),
		IdleTimeout:        milliToDuration(cfg.GetIdleTimeout()),
		IdleCheckFrequency: milliToDuration(cfg.GetIdleCheckFrequency()),
	})
	return &redisClient{
		cli: cli,
	}
}

func newFailoverClient(cfg *cacheProxy.CacheOpt) *redisClient {
	cli := redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:       cfg.GetFailOver().GetMasterName(),
		SentinelAddrs:    cfg.GetFailOver().GetSentinels(),
		SentinelPassword: cfg.GetFailOver().GetSentinelPassword(),

		Password:        cfg.GetPassword(),
		MaxRetries:      int(cfg.GetMaxRetries()),
		MinRetryBackoff: milliToDuration(cfg.GetMinRetryBackoff()),
		MaxRetryBackoff: milliToDuration(cfg.GetMaxRetryBackoff()),
		DialTimeout:     milliToDuration(cfg.GetDialTimeout()),
		ReadTimeout:     milliToDuration(cfg.GetReadTimeout()),
		WriteTimeout:    milliToDuration(cfg.GetWriteTimeout()),
		PoolSize:        int(cfg.GetPoolSize()),
		MinIdleConns:    int(cfg.GetMinIdleConns()),

		MaxConnAge:         milliToDuration(cfg.GetMaxConnAge()),
		PoolTimeout:        milliToDuration(cfg.GetPoolTimeout()),
		IdleTimeout:        milliToDuration(cfg.GetIdleTimeout()),
		IdleCheckFrequency: milliToDuration(cfg.GetIdleCheckFrequency()),
	})
	return &redisClient{
		cli: cli,
	}
}

func newClusterClient(cfg *cacheProxy.CacheOpt) *redisClient {
	cli := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: cfg.GetCluster().GetAddr(),

		Password:        cfg.GetPassword(),
		MaxRetries:      int(cfg.GetMaxRetries()),
		MinRetryBackoff: milliToDuration(cfg.GetMinRetryBackoff()),
		MaxRetryBackoff: milliToDuration(cfg.GetMaxRetryBackoff()),
		DialTimeout:     milliToDuration(cfg.GetDialTimeout()),
		ReadTimeout:     milliToDuration(cfg.GetReadTimeout()),
		WriteTimeout:    milliToDuration(cfg.GetWriteTimeout()),
		PoolSize:        int(cfg.GetPoolSize()),
		MinIdleConns:    int(cfg.GetMinIdleConns()),

		MaxConnAge:         milliToDuration(cfg.GetMaxConnAge()),
		PoolTimeout:        milliToDuration(cfg.GetPoolTimeout()),
		IdleTimeout:        milliToDuration(cfg.GetIdleTimeout()),
		IdleCheckFrequency: milliToDuration(cfg.GetIdleCheckFrequency()),
	})
	return &redisClient{
		cli: cli,
	}
}

func milliToDuration(milli int64) time.Duration {
	return time.Duration(milli) * time.Millisecond
}

func (c *redisClient) HashSet(ctx context.Context, key string, fields map[string]interface{}, ttl int) error {
	var err = c.cli.HMSet(ctx, key, fields).Err()
	if err != nil {
		return err
	}
	err = c.cli.Expire(ctx, key, time.Duration(ttl)*time.Second).Err()
	if err != nil {
		return err
	}
	return nil
}

func (c *redisClient) HashGet(ctx context.Context, key string, field string) ([]byte, error) {
	var ret = c.cli.HGet(ctx, key, field)
	var buf, err = ret.Bytes()
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func (c *redisClient) HashGetAll(ctx context.Context, key string) (map[string]string, error) {
	var ret = c.cli.HGetAll(ctx, key)
	var fields, err = ret.Result()
	if err != nil {
		return nil, err
	}
	return fields, nil
}

func (c *redisClient) PHashGetAll(ctx context.Context, key ...string) (map[string]string, error) {
	var (
		data = make(map[string]string)
		pipe = c.cli.Pipeline()
		cmdS = make(map[string]redis.Cmder)
		err  error
	)

	for _, k := range key {
		cmdS[k] = pipe.HGetAll(ctx, k)
	}
	_, err = pipe.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("redis pipine exec fail. %s", err.Error())
	}

	for k, cmd := range cmdS {
		sm, err := cmd.(*redis.StringStringMapCmd).Result()
		if err != nil {
			return nil, fmt.Errorf("parse (key=%s) cache data fail. %s", k, err.Error())
		}
		for kk, vv := range sm {
			data[kk] = vv
		}
	}
	return data, nil
}
