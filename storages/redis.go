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

func (r *Redis) ListPush(method bool, key string, values [][]byte) error {
	var err error
	in := make([]interface{}, 0, len(values))
	for i := 0; i < len(values); i++ {
		in = append(in, values[i])
	}
	if len(in) <= 0 {
		return nil
	}
	if method {
		err = r.conn.LPush(key, in...).Err()
	} else {
		err = r.conn.RPush(key, in...).Err()
	}
	if err != nil {
		r.logger.Error(err.Error())
		return err
	}
	return nil
}

func (r *Redis) ListPop(method bool, key string) ([]byte, error) {
	var ret *redis.StringCmd
	if method {
		ret = r.conn.LPop(key)
	} else {
		ret = r.conn.RPop(key)
	}
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

func (r *Redis) ListLength(key string) (int64, error) {
	length, err := r.conn.LLen(key).Result()
	if err != nil {
		r.logger.Error(err.Error())
		return 0, err
	}
	return length, nil
}

func (r *Redis) ListDelete(key string, value []byte) error {
	err := r.conn.LRem(key, 0, value).Err()
	if err != nil {
		r.logger.Error(err.Error())
		return err
	}
	return nil
}

func (r *Redis) HashSet(key string, fields map[string][]byte) error {
	in := make(map[string]interface{})
	for k, v :=range fields {
		in[k] = v
	}
	if len(in) <= 0 {
		return nil
	}
	err := r.conn.HMSet(key, in).Err()
	if err != nil {
		r.logger.Error(err.Error())
		return err
	}
	return nil
}

func (r *Redis) HashGet(key string, field string) ([]byte, error) {
	ret := r.conn.HGet(key, field)
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

func (r *Redis) HashDelete(key string, fields ...string) error {
	err := r.conn.HDel(key, fields...).Err()
	if err != nil {
		r.logger.Error(err.Error())
		return err
	}
	return nil
}
