//author: richard
package caches

import (
	"github.com/advancevillage/3rd/logs"
	"github.com/advancevillage/3rd/storages"
	"github.com/go-redis/redis"
)

const (
	CacheSeparator = "-"
)

type Cache interface {
	DeleteCache(key ...string) error
	QueryCache(key  string, timeout int) ([]byte, error)
	UpdateCache(key string, body []byte, timeout int) error
	CreateCache(key string, body []byte, timeout int) error
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