//author: richard
package caches

import (
	"fmt"
	"github.com/advancevillage/3rd/logs"
	"github.com/go-redis/redis"
)

func NewRedisMessage(host string, port int, auth string, schema int, logger logs.Logs) (IMessage, error) {
	c := &Message{}
	c.logger = logger
	c.schema = schema
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
	err = c.conn.Do("config", "set", "notify-keyspace-events", "KEA").Err()
	if err != nil {
		c.logger.Emergency(err.Error())
		return nil, err
	}
	return c, nil
}

func (s *Message) Publish(channel string, data []byte) error {
	err := s.conn.Publish(channel, data).Err()
	if err != nil {
		s.logger.Error(err.Error())
		return err
	}
	return nil
}

func (s *Message) Subscribe(key string, f func(string, []byte) error) error {
	sub := s.conn.Subscribe(key)
	_, err := sub.Receive()
	if err != nil {
		s.logger.Error(err.Error())
		return err
	}
	ch := sub.Channel()
	for message :=range ch {
		err = f(key, []byte(message.Payload))
		if err != nil {
			s.logger.Error(err.Error())
			break
		} else {
			continue
		}
	}
	return nil
}

func (s *Message) KeySpace(key string, f func(string, []byte) error) error {
	channel := fmt.Sprintf("__%s@%d__:%s", "keyspace", s.schema, key)
	sub := s.conn.Subscribe(channel)
	_, err := sub.Receive()
	if err != nil {
		s.logger.Error(err.Error())
		return err
	}
	ch := sub.Channel()
	for message :=range ch {
		err = f(key, []byte(message.Payload))
		if err != nil {
			s.logger.Error(err.Error())
			break
		} else {
			continue
		}
	}
	return nil
}
