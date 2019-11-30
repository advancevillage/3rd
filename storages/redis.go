//author: richard
package storages

import (
	"3rd/logs"
	"fmt"
	"github.com/go-redis/redis"
	"time"
)

func NewRedis(host string, port int, auth string, schema int, logger logs.Logs) (*Redis, error) {
	r := &Redis{}
	r.host = host
	r.port = port
	r.schema = schema
	r.auth = auth
	r.logger = logger
	r.conn = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", r.host, r.port),
		Password: r.auth,
		DB:  r.schema,
	})
	_, err := r.conn.Ping().Result()
	if err != nil {
		r.logger.Emergency(err.Error())
		return nil, err
	}
	return r, nil
}

//@link: http://redisdoc.com/string/setex.html
//SET key value EX XX|NX
func (r *Redis) StrSet(key string, value []byte, timeout int) error {
	err := r.conn.Set(key, value, time.Duration(timeout) * time.Second).Err()
	if err != nil {
		r.logger.Error(err.Error())
		return err
	}
	return nil
}

func (r *Redis) StrGet(key string) ([]byte, error) {
	ret := r.conn.Get(key)
	if ret == nil {
		return nil, ErrorKeyNotExist
	}
	buf, err := ret.Bytes()
	if err != nil {
		r.logger.Error(err.Error())
		return nil, ErrorKeyNotExist
	}
	return buf, nil
}

func (r *Redis) LPush(key string, values ...[]byte) error {
	err := r.conn.LPush(key, values).Err()
	if err != nil {
		r.logger.Error(err.Error())
		return err
	}
	return nil
}

func (r *Redis) RPush(key string, values ...[]byte) error {
	err := r.conn.RPush(key, values).Err()
	if err != nil {
		r.logger.Error(err.Error())
		return err
	}
	return nil
}

func (r *Redis) LPop(key string) ([]byte, error) {
	ret := r.conn.LPop(key)
	if ret == nil {
		return nil, ErrorKeyNotExist
	}
	buf, err := ret.Bytes()
	if err != nil {
		r.logger.Error(err.Error())
		return nil, ErrorKeyNotExist
	}
	return buf, nil
}

func (r *Redis) RPop(key string) ([]byte, error) {
	ret := r.conn.RPop(key)
	if ret == nil {
		return nil, ErrorKeyNotExist
	}
	buf, err := ret.Bytes()
	if err != nil {
		r.logger.Error(err.Error())
		return nil, ErrorKeyNotExist
	}
	return buf, nil
}

func (r *Redis) LLen(key string) (int64, error) {
	length, err := r.conn.LLen(key).Result()
	if err != nil {
		r.logger.Error(err.Error())
		return 0, err
	}
	return length, nil
}
