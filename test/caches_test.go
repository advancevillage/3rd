//author: richard
package test

import (
	"encoding/json"
	"github.com/advancevillage/3rd/caches"
	"github.com/advancevillage/3rd/logs"
	"github.com/advancevillage/3rd/storages"
	"log"
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
	cache, err := caches.NewRedisCache("192.168.1.101", 6379, "", 0, logger, leveldb)
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

func TestRedisStorage (t *testing.T) {
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
	cache, err := caches.NewRedisStorage("192.168.1.101", 6379, "", 0, logger)
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

func TestIMessage(t *testing.T) {
	logger, err := logs.NewTxtLogger("cache.log", 64, 4)
	if err != nil {
		t.Error(err.Error())
		return
	}
	message, err := caches.NewRedisMessage("localhost", 6379, "", 0, logger)
	if err != nil {
		t.Error(err.Error())
		return
	}
	channel := "richard"
	//test := struct {
	//	KeySecret string `json:"keySecret"`
	//	RobotId   string `json:"robotId"`
	//}{
	//	KeySecret: utils.UUID(),
	//	RobotId: utils.UUID(),
	//}
	//buf, err := json.Marshal(test)
	//if err != nil {
	//	t.Error(err.Error())
	//	return
	//}
	f := func(key string, buf []byte) error {
		log.Println(key, string(buf))
		return nil
	}
	go func() {
		err = message.KeySpace("kelly", f)
		if err != nil {
			t.Error(err.Error())
		}
	}()
	err = message.KeySpace(channel, f)
	if err != nil {
		t.Error(err.Error())
	}
	//err = message.Publish(channel, buf)
	//if err != nil {
	//	t.Error(err.Error())
	//}
}