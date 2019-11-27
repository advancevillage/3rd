//author: richard
package caches

import (
	"3rd/logs"
	"3rd/storages"
)

type Cache interface {
	UpdateCache(key string, body []byte) error
	CreateCache(key string, body []byte) error
	QueryCache(key  string) ([]byte, error)
	DeleteCache(key string) error
}

type Redis struct {
	logger   logs.Logs
	storage  storages.Storage
}