//author: richard
package caches

import (
	"3rd/logs"
	"3rd/storages"
)

const (
	CacheSeparator = "-"
)

type CacheOptions struct {
	Timeout int64
}

type Cache interface {
	UpdateCache(key string, body []byte, options *CacheOptions) error
	CreateCache(key string, body []byte, options *CacheOptions) error
	QueryCache(key  string, options *CacheOptions) ([]byte, error)
	DeleteCache(key string) error
}

type Redis struct {
	logger   logs.Logs
	storage  storages.Storage
}