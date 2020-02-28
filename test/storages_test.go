//author: richard
package test

import (
	"encoding/json"
	"github.com/advancevillage/3rd/logs"
	"github.com/advancevillage/3rd/storages"
	"github.com/advancevillage/3rd/utils"
	"strconv"
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
	err = redis.CreateStorage("richard", []byte("ShowU, A More Beautiful Self"))
	if err != nil {
		t.Error(err.Error())
	}
	value, err := redis.QueryStorage("richard")
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
		err = redis.CreateStorage(key, value)
		if err != nil {
			b.Error(err.Error())
		}
		_, err := redis.QueryStorage(key)
		if err != nil {
			b.Error(err.Error())
			return
		}
	}
}

func TestStorageV2(t *testing.T) {
	logger, err := logs.NewTxtLogger("storage.log", 512, 4)
	if err != nil {
		t.Error(err.Error())
	}
	redis, err := storages.NewRedis("localhost", 6379, "", 0, logger)
	if err != nil {
		t.Error(err.Error())
		return
	}
	es7, err := storages.NewTES([]string{"http://localhost:9200"}, logger)
	if err != nil {
		t.Error(err.Error())
		return
	}
	db, err := storages.NewLevelDB("database", logger)
	if err != nil {
		t.Error(err.Error())
		return
	}
	index := "richard"
	key := "kelly"
	object := struct {
		Id  int64 	`json:"id,omitempty"`
		Name string `json:"name,omitempty"`
		Age  int 	`json:"age,omitempty"`
	}{
		Id: utils.SnowFlakeId(),
		Name: utils.RandsString(12),
		Age: utils.RandsInt(100),
	}
	body, err := json.Marshal(object)
	if err != nil {
		t.Error(err.Error())
		return
	}
	err = redis.CreateStorageV2(index, key, body)
	if err != nil {
		t.Error(err.Error())
		return
	}
	err = es7.CreateStorageV2(index, key, body)
	if err != nil {
		t.Error(err.Error())
		return
	}
	err = db.CreateStorageV2(index, key, body)
	if err != nil {
		t.Error(err.Error())
		return
	}
	buf, err := redis.QueryStorageV2(index, key)
	if err != nil {
		t.Error(err.Error())
		return
	}
	t.Log(string(buf))
	buf, err = es7.QueryStorageV2(index, key)
	if err != nil {
		t.Error(err.Error())
		return
	}
	t.Log(string(buf))
	buf, err = db.QueryStorageV2(index, key)
	if err != nil {
		t.Error(err.Error())
		return
	}
	t.Log(string(buf))
}

func TestAwsEs(t *testing.T) {
	logger, err := logs.NewTxtLogger("storage.log", 512, 4)
	if err != nil {
		t.Error(err.Error())
	}
	index := "richard"
	key := "kelly"
	object := struct {
		Id  int64 	`json:"id,omitempty"`
		Name string `json:"name,omitempty"`
		Age  int 	`json:"age,omitempty"`
	}{
		Id: utils.SnowFlakeId(),
		Name: utils.RandsString(12),
		Age: utils.RandsInt(100),
	}
	body, err := json.Marshal(object)
	if err != nil {
		t.Error(err.Error())
		return
	}
	awsEs, err := storages.NewAwsES("AKIA5MGTVEAKFPBRN2FF", "RoIqxVnIxQkfb9Xdsncj45MfZH6bnYBc8+KxiE14", "ap-southeast-1", "https://search-test-4v5zjk23vcg2wbrt6noh3lmxhm.ap-southeast-1.es.amazonaws.com", logger)
	if err != nil {
		t.Error(err.Error())
	}
	err = awsEs.CreateStorageV2(index, key, body)
	if err != nil {
		t.Error(err.Error())
	}
}

func TestMongoDB(t *testing.T) {
	logger, err := logs.NewTxtLogger("mongo.log", 512, 4)
	if err != nil {
		t.Error(err.Error())
	}
	mgo, err := storages.NewMongoDB("mongodb://admin:password@localhost:27017", logger)
	if err != nil {
		t.Error(err.Error())
	}
	index := "carts"
	key := "233349905152741376"
	where := make(map[string]interface{})
	sort  := make(map[string]interface{})
	sort["createTime"] = -1
	where["createTime"] =  map[string]interface{}{"$gt": 0}
	items, total, err := mgo.SearchStorageV2Exd(index, key, where, 99, 0, sort)
	if err != nil {
		t.Error(err.Error())
		return
	}
	t.Log(items, total)

}

func TestMongoExd(t *testing.T) {
	logger, err := logs.NewTxtLogger("mongo.log", 512, 4)
	if err != nil {
		t.Error(err.Error())
	}
	mgo, err := storages.NewMongoDB("mongodb://admin:password@localhost:27017", logger)
	if err != nil {
		t.Error(err.Error())
	}
	index := "richard"
	key := "111111"
	object := struct {
		Id   string `json:"id,omitempty"`
		Name string `json:"name,omitempty"`
		Age  int 	`json:"age,omitempty"`
	}{
		Id: utils.SnowFlakeIdString(),
		Name: utils.RandsString(12),
		Age: utils.RandsInt(100),
	}
	buf, err := json.Marshal(object)
	if err != nil {
		t.Error(err.Error())
		return
	}
	err = mgo.CreateStorageV2Exd(index, key, object.Id, buf)
	if err != nil {
		t.Error(err.Error())
		return
	}
	return
}

func TestMongoDB_QueryV3(t *testing.T) {
	logger := logs.NewStdLogger()
	mgo, err := storages.NewMongoDB("mongodb://admin:password@localhost:27017", logger)
	if err != nil {
		t.Error(err.Error())
		return
	}
	index := "categories"
	where := make(map[string]interface{})
	//where["brandStatus"] = 0x701
	sort := make(map[string]interface{})
	//sort["brandCreateTime"] = -1
	sort["createTime"] = 1

	body, total, err := mgo.QueryStorageV3(index, where, 100, 0, sort)
	if err != nil {
		t.Error(err.Error())
		return
	}
	for i := range body {
		t.Log(total, string(body[i]))
	}
}

func TestMongoDBStorageInterface(t *testing.T) {
	logger, err := logs.NewTxtLogger("mongo.log", 512, 4)
	if err != nil {
		t.Error(err.Error())
	}
	mgo, err := storages.NewMongoDB("mongodb://localhost:27017", logger)
	if err != nil {
		t.Error(err.Error())
	}
	index := "categories"
	object := struct {
		Id   string `json:"id,omitempty"`
		Name string `json:"name,omitempty"`
		Age  int 	`json:"age,omitempty"`
	}{
		Id: utils.SnowFlakeIdString(),
		Name: utils.RandsString(12),
		Age: utils.RandsInt(100),
	}
	body, err := json.Marshal(object)
	if err != nil {
		t.Error(err.Error())
		return
	}
	err = mgo.CreateStorageV2(index, object.Id, body)
	if err != nil {
		t.Error(err.Error())
		return
	}
	body, err = mgo.QueryStorageV2(index, object.Id)
	if err != nil {
		t.Error(err.Error())
		return
	}
	object.Name = "test"
	object.Age  = 200
	body, err = json.Marshal(object)
	if err != nil {
		t.Error(err.Error())
		return
	}
	err = mgo.UpdateStorageV2(index, object.Id, body)
	if err != nil {
		t.Error(err.Error())
		return
	}
	err = mgo.DeleteStorageV2(index, object.Id)
	if err != nil {
		t.Error(err.Error())
		return
	}
}

func TestMongoDBExd(t *testing.T) {
	mgo, err := storages.NewMongoDB("mongodb://admin:password@localhost:27017", logs.NewStdLogger())
	if err != nil {
		t.Error(err.Error())
		return
	}
	stock := struct {
		Id  int `json:"id"`
		GoodsId string `json:"goodsId,omitempty"`
		ColorId string `json:"colorId,omitempty"`
		SizeId  string `json:"sizeId,omitempty"`
		SizeValue string `json:"sizeValue,omitempty"`
		ColorName string `json:"colorName,omitempty"`
		Count   int    `json:"count,omitempty"`
		Version int    `json:"version,omitempty"`
	}{
		Id: 13,
		GoodsId: "0000000000000000",
		ColorId: "0000000000000001",
		SizeId: "00000000000000002",
		SizeValue: "38",
		ColorName: "red",
		Count: 100,
		Version: 3,
	}
	buf, err := json.Marshal(stock)
	if err != nil {
		t.Error(err.Error())
		return
	}

	index := "stocks"

	err = mgo.CreateStorageV2Exd(index, stock.GoodsId, strconv.Itoa(stock.Id), buf)
	if err != nil {
		t.Error(err.Error())
		return
	}

	where := make(map[string]interface{})
	where["goodsId"] = "0000000000000000"
	where["sizeId"]  = "00000000000000002"
	where["colorId"] = "0000000000000001"
	where["version"] = 3

	stock.Version += 1
	stock.Id       = 10

	buf, err = json.Marshal(stock)
	if err != nil {
		t.Error(err.Error())
		return
	}

	err = mgo.UpdateStorageV2Exd(index, stock.GoodsId, where, strconv.Itoa(stock.Id), buf)
	if err != nil {
		t.Error(err.Error())
		return
	}

	stock.Version += 1
	stock.Count   -= 60

	buf, err = json.Marshal(stock)
	if err != nil {
		t.Error(err.Error())
		return
	}

	err = mgo.UpdateStorageV2Exd(index, stock.GoodsId, where, strconv.Itoa(stock.Id), buf)
	if err != nil {
		t.Error(err.Error())
		return
	}

}

//goos: darwin
//goarch: amd64
//pkg: github.com/advancevillage/3rd/test
//BenchmarkAwsEs-8   	       1	1347722079 ns/op
//PASS
func BenchmarkAwsEs(b *testing.B) {
	logger, err := logs.NewTxtLogger("storage.log", 512, 4)
	if err != nil {
		b.Error(err.Error())
	}
	object := struct {
		Id  int64 	`json:"id,omitempty"`
		Name string `json:"name,omitempty"`
		Age  int 	`json:"age,omitempty"`
	}{}
	awsEs, err := storages.NewAwsES("AKIA5MGTVEAKFPBRN2FF", "RoIqxVnIxQkfb9Xdsncj45MfZH6bnYBc8+KxiE14", "ap-southeast-1", "https://search-test-4v5zjk23vcg2wbrt6noh3lmxhm.ap-southeast-1.es.amazonaws.com", logger)
	for i := 0; i < b.N; i++ {
		index := "richard"
		key := utils.RandsString(10)
		object.Id = utils.SnowFlakeId()
		object.Name = utils.RandsString(10)
		object.Age = utils.RandsInt(100)
		body, err := json.Marshal(object)
		if err != nil {
			b.Error(err.Error())
			continue
		}
		err = awsEs.CreateStorageV2(index, key, body)
		if err != nil {
			b.Error(err.Error())
		}
	}
}

func BenchmarkMongo(b *testing.B) {
	logger, err := logs.NewTxtLogger("storage.log", 512, 4)
	if err != nil {
		b.Error(err.Error())
	}
	object := struct {
		Id  int64 	`json:"id,omitempty"`
		Name string `json:"name,omitempty"`
		Age  int 	`json:"age,omitempty"`
	}{}
	mgo, err := storages.NewMongoDB("mongodb://localhost:27017", logger)
	if err != nil {
		b.Error(err.Error())
	}
	index := "richard"
	key := "test"
	for i := 0; i < b.N; i++ {
		object.Id = utils.SnowFlakeId()
		object.Name = utils.RandsString(10)
		object.Age = utils.RandsInt(100)
		err = mgo.CreateStorageV2(index, key, []byte("11"))
		if err != nil {
			b.Error(err.Error())
		}
	}
}

//goos: darwin
//goarch: amd64
//pkg: github.com/advancevillage/3rd/test
//BenchmarkMongoDBStorageInterface_CreateV2-8   	    1042	   1025497 ns/op
func BenchmarkMongoDBStorageInterface_CreateV2(b *testing.B) {
	logger, err := logs.NewTxtLogger("storage.log", 512, 4)
	if err != nil {
		b.Error(err.Error())
	}
	object := struct {
		Id   string 	`json:"id,omitempty"`
		Name string `json:"name,omitempty"`
		Age  int 	`json:"age,omitempty"`
	}{}
	mgo, err := storages.NewMongoDB("mongodb://localhost:27017", logger)
	if err != nil {
		b.Error(err.Error())
	}
	index := "testing"
	failCount := 0
	for i := 0; i < b.N; i++ {
		object.Id = utils.SnowFlakeIdString()
		object.Name = utils.RandsString(1000)
		object.Age = utils.RandsInt(10000)
		body, err := json.Marshal(object)
		if err != nil {
			b.Error(err.Error())
			failCount++
			continue
		}
		err = mgo.CreateStorageV2(index, object.Id, body)
		if err != nil {
			b.Error(err.Error())
			failCount++
			continue
		}
	}
	b.Log(failCount)
}

func BenchmarkMongoDBStorageInterface_QueryV2(b *testing.B) {
	logger, err := logs.NewTxtLogger("storage.log", 512, 4)
	if err != nil {
		b.Error(err.Error())
	}
	mgo, err := storages.NewMongoDB("mongodb://localhost:27017", logger)
	if err != nil {
		b.Error(err.Error())
	}
	index := "testing"
	key := "215229069808111616"
	failCount := 0
	for i := 0; i < b.N; i++ {
		_, err := mgo.QueryStorageV2(index, key)
		if err != nil {
			b.Error(err.Error())
			failCount++
			continue
		}

	}
	b.Log(failCount)
}



