//author: richard
package caches

import (
	"3rd/logs"
	"3rd/storages"
)

func NewRedis(logger logs.Logs, storage storages.Storage) *Redis {
	c := &Redis{}
	c.logger = logger
	c.storage = storage
	return c
}
