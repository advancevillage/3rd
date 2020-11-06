//author: richard
package caches

import (
	"errors"

	"github.com/advancevillage/3rd/logs"
	"github.com/advancevillage/3rd/storages"

	"github.com/go-redis/redis/v8"
)

var (
	ErrorKeyNotExist = errors.New("key not exist")
)

const (
	CacheSeparator = "-"
)

type ICache interface {
	DeleteCache(key ...string) error
	QueryCache(key string, timeout int) ([]byte, error)
	UpdateCache(key string, body []byte, timeout int) error
	CreateCache(key string, body []byte, timeout int) error

	CreateCacheV2(index string, key string, body []byte) error
	UpdateCacheV2(index string, key string, body []byte) error
	QueryCacheV2(index string, key string) ([]byte, error)
	DeleteCacheV2(index string, key ...string) error
}

type IMessage interface {
	Publish(channel string, data []byte) error
	KeySpace(key string, f func(string, []byte) error) error
	Subscribe(channel string, f func(string, []byte) error) error
}

type Cache struct {
	conn    *redis.Client
	logger  logs.Logs
	storage storages.Storage
}

type Storage struct {
	conn   *redis.Client
	logger logs.Logs
}

type Message struct {
	conn   *redis.Client
	logger logs.Logs
	schema int
}
