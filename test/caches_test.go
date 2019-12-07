//author: richard
package test

import (
	"encoding/json"
	"github.com/advancevillage/3rd/caches"
	"github.com/advancevillage/3rd/logs"
	"github.com/advancevillage/3rd/storages"
	"testing"
	"time"
)

func TestRedisCache (t *testing.T) {
	logger, err := logs.NewTxtLogger("cache.log", 512, 4)
	if err != nil {
		t.Error(err.Error())
		return
	}
	leveldb, err := storages.NewLevelDB("database", logger)
	if err != nil {
		t.Error(err.Error())
		return
	}
	goods := struct {
		GoodName  string	`json:"good_name"`
		GoodPrice float32	`json:"good_price"`
		GoodStock int		`json:"good_stock"`
	}{
		GoodName: "clothes",
		GoodPrice: 99.9,
		GoodStock: 100,
	}
	buf, err := json.Marshal(goods)
	if err != nil {
		t.Error(err.Error())
		return

	}
	key := "richard"
	err = leveldb.CreateStorage(key, buf)
	if err != nil {
		t.Error(err.Error())
		return
	}
	cache, err := caches.NewRedis("192.168.1.101", 6379, "", 0, logger, leveldb)
	if err != nil {
		t.Error(err.Error())
		return
	}
	for i := 0; i < 1; i++ {
		_, err := cache.QueryCache(key, 180)
		if err != nil {
			t.Error(err.Error())
			continue
		}
		time.Sleep(3 * time.Second)
	}
}
