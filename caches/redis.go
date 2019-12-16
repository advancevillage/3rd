//author: richard
package caches

import (
	"fmt"
	"github.com/advancevillage/3rd/logs"
	"github.com/advancevillage/3rd/storages"
	"github.com/go-redis/redis"
	"time"
)

func NewRedis(host string, port int, auth string, schema int, logger logs.Logs, storage storages.Storage) (*Redis, error) {
	c := &Redis{}
	c.host = host
	c.port = port
	c.schema = schema
	c.auth = auth
	c.logger = logger
	c.storage = storage
	c.conn = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", c.host, c.port),
		Password: c.auth,
		DB:  c.schema,
	})
	_, err := c.conn.Ping().Result()
	if err != nil {
		c.logger.Emergency(err.Error())
		return nil, err
	}
	return c, nil
}

func (c *Redis) CreateCache(key string, body []byte, timeout int) error {
	err := c.conn.SetNX(key, body, time.Duration(timeout) * time.Second).Err()
	if err != nil {
		c.logger.Error(err.Error())
		return err
	}
	return nil
}

func (c *Redis) UpdateCache(key string, body []byte, timeout int) error {
	err := c.conn.SetXX(key, body, time.Duration(timeout) * time.Second).Err()
	if err != nil {
		c.logger.Error(err.Error())
		return err
	}
	return nil
}

func (c *Redis) QueryCache(key  string, timeout int) ([]byte, error) {
	ret := c.conn.Get(key)
	buf, err := ret.Bytes()
	if err != nil {
		//缓存层不存在,查询存储层
		value, err := c.storage.QueryStorage(key)
		if err != nil {
			return nil, storages.ErrorKeyNotExist
		}
		//更新缓存层
		go func () { _ = c.CreateCache(key, value, timeout)}()
		return value, nil
	}
	return buf, nil
}

func (c *Redis) DeleteCache(key ...string) error {
	err := c.conn.Del(key...).Err()
	if err != nil {
		c.logger.Error(err.Error())
		return err
	}
	return nil
}

func (c *Redis) CreateCacheV2(index string, key string, body []byte) error {
	var fields = make(map[string][]byte)
	fields[key] = body
	return c.HashSet(index, fields)
}

func (c *Redis) UpdateCacheV2(index string, key string, body []byte) error {
	var fields = make(map[string][]byte)
	fields[key] = body
	return c.HashSet(index, fields)
}

func (c *Redis) QueryCacheV2(index string, key  string) ([]byte, error) {
	return c.HashGet(index, key)
}

func (c *Redis) DeleteCacheV2(index string, key ...string) error {
	return c.HashDelete(index, key ...)
}

func (c *Redis) HashSet(key string, fields map[string][]byte) error {
	in := make(map[string]interface{})
	for k, v :=range fields {
		in[k] = v
	}
	if len(in) <= 0 {
		return nil
	}
	err := c.conn.HMSet(key, in).Err()
	if err != nil {
		c.logger.Error(err.Error())
		return err
	}
	return nil
}

func (c *Redis) HashGet(key string, field string) ([]byte, error) {
	ret := c.conn.HGet(key, field)
	buf, err := ret.Bytes()
	if err != nil {
		c.logger.Error(err.Error())
		body, err := c.storage.QueryStorageV2(key, field)
		if err != nil {
			return nil, ErrorKeyNotExist
		}
		_ = c.storage.CreateStorageV2(key, field, body)
		return body, nil
	}
	return buf, nil
}

func (c *Redis) HashDelete(key string, fields ...string) error {
	err := c.conn.HDel(key, fields...).Err()
	if err != nil {
		c.logger.Error(err.Error())
		return err
	}
	return nil
}