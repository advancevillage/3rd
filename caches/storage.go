//author: richard
package caches

import (
	"fmt"
	"github.com/advancevillage/3rd/logs"
	"github.com/advancevillage/3rd/storages"
	"github.com/go-redis/redis"
	"time"
)

func NewRedisStorage(host string, port int, auth string, schema int, logger logs.Logs) (ICache, error) {
	c := &Storage{}
	c.logger = logger
	c.conn = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", host, port),
		Password: auth,
		DB:  schema,
	})
	_, err := c.conn.Ping().Result()
	if err != nil {
		c.logger.Emergency(err.Error())
		return nil, err
	}
	return c, nil
}

func (c *Storage) CreateCache(key string, body []byte, timeout int) error {
	err := c.conn.SetNX(key, body, time.Duration(timeout) * time.Second).Err()
	if err != nil {
		c.logger.Error(err.Error())
		return err
	}
	return nil
}

func (c *Storage) UpdateCache(key string, body []byte, timeout int) error {
	err := c.conn.SetXX(key, body, time.Duration(timeout) * time.Second).Err()
	if err != nil {
		c.logger.Error(err.Error())
		return err
	}
	return nil
}

func (c *Storage) QueryCache(key  string, timeout int) ([]byte, error) {
	ret := c.conn.Get(key)
	buf, err := ret.Bytes()
	if err != nil {
		return nil, storages.ErrorKeyNotExist
	}
	return buf, nil
}

func (c *Storage) DeleteCache(key ...string) error {
	err := c.conn.Del(key...).Err()
	if err != nil {
		c.logger.Error(err.Error())
		return err
	}
	return nil
}

func (c *Storage) CreateCacheV2(index string, key string, body []byte) error {
	var fields = make(map[string][]byte)
	fields[key] = body
	return c.HashSet(index, fields)
}

func (c *Storage) UpdateCacheV2(index string, key string, body []byte) error {
	var fields = make(map[string][]byte)
	fields[key] = body
	return c.HashSet(index, fields)
}

func (c *Storage) QueryCacheV2(index string, key  string) ([]byte, error) {
	return c.HashGet(index, key)
}

func (c *Storage) DeleteCacheV2(index string, key ...string) error {
	return c.HashDelete(index, key ...)
}

func (c *Storage) HashSet(key string, fields map[string][]byte) error {
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

func (c *Storage) HashGet(key string, field string) ([]byte, error) {
	ret := c.conn.HGet(key, field)
	buf, err := ret.Bytes()
	if err != nil {
		return nil, ErrorKeyNotExist
	}
	return buf, nil
}

func (c *Storage) HashDelete(key string, fields ...string) error {
	err := c.conn.HDel(key, fields...).Err()
	if err != nil {
		c.logger.Error(err.Error())
		return err
	}
	return nil
}

