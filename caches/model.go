//author: richard
package caches

import (
	"errors"
	"github.com/advancevillage/3rd/logs"
	"github.com/advancevillage/3rd/storages"
	"github.com/go-redis/redis"
)

var (
	ErrorKeyNotExist = errors.New("key not exist")
)

const (
	CacheSeparator = "-"
)

type Cache interface {
	DeleteCache(key ...string) error
	QueryCache(key  string, timeout int) ([]byte, error)
	UpdateCache(key string, body []byte, timeout int) error
	CreateCache(key string, body []byte, timeout int) error

	CreateCacheV2(index string, key string, body []byte) error
	UpdateCacheV2(index string, key string, body []byte) error
	QueryCacheV2(index string, key  string) ([]byte, error)
	DeleteCacheV2(index string, key ...string) error
}

type Redis struct {
	host   string
	port   int
	auth   string
	schema int
	conn     *redis.Client
	logger   logs.Logs
	storage  storages.Storage
}