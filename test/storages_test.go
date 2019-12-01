//author: richard
package test

import (
	"3rd/logs"
	"3rd/storages"
	"testing"
)

func TestStorageRedis(t *testing.T) {
	logger, err := logs.NewTxtLogger("storage.log", 512, 4)
	if err != nil {
		t.Error(err.Error())
	}
	redis, err := storages.NewRedis("192.168.1.101", 6379, "", 0, logger)
	if err != nil {
		t.Error(err.Error())
		return
	}
	err = redis.StrSet("richard", []byte("ShowU, A More Beautiful Self"), 600)
	if err != nil {
		t.Error(err.Error())
	}
	value, err := redis.StrGet("richard")
	if err != nil {
		t.Error(err.Error())
		return
	}
	t.Log(value)
}

func BenchmarkStorageRedis(b *testing.B) {
	logger, err := logs.NewTxtLogger("storage.log", 512, 4)
	if err != nil {
		b.Error(err.Error())
	}
	redis, err := storages.NewRedis("192.168.1.101", 6379, "", 0, logger)
	if err != nil {
		b.Error(err.Error())
		return
	}
	key := "richard"
	value := []byte("ShowU, A More Beautiful Self")
	for i := 0; i < b.N; i++ {
		err = redis.StrSet(key, value, 600)
		if err != nil {
			b.Error(err.Error())
		}
		_, err := redis.StrGet(key)
		if err != nil {
			b.Error(err.Error())
			return
		}
	}
}

func TestStorageRedis_List(t *testing.T) {
	logger, err := logs.NewTxtLogger("storage.log", 512, 4)
	if err != nil {
		t.Error(err.Error())
	}
	redis, err := storages.NewRedis("192.168.1.101", 6379, "", 0, logger)
	if err != nil {
		t.Error(err.Error())
		return
	}
	values := [][]byte {
		[]byte("richard"),
		[]byte("sun"),
	}
	key := "kelly"
	err = redis.ListPush(true, key, values)
	if err != nil {
		t.Error(err.Error())
	}
	_, err = redis.ListPop(false, key)
	if err != nil {
		t.Error(err.Error())
	}
	err = redis.ListDelete(key, values[0])
	if err != nil {
		t.Error(err.Error())
	}
	err = redis.ListPush(false, key, nil)
	if err != nil {
		t.Error(err.Error())
	}
}

