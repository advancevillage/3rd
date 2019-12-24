//author: richard
package test

import (
	"encoding/json"
	"github.com/advancevillage/3rd/logs"
	"github.com/advancevillage/3rd/storages"
	"github.com/advancevillage/3rd/utils"
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
	mgo, err := storages.NewMongoDB("mongodb://localhost:27017", logger)
	if err != nil {
		t.Error(err.Error())
	}
	index := "richard"
	key := "666666"
	object := struct {
		Id  int64 	`json:"id,omitempty"`
		Name string `json:"name,omitempty"`
		Age  int 	`json:"age,omitempty"`
	}{
		Id: utils.SnowFlakeId(),
		Name: utils.RandsString(12),
		Age: utils.RandsInt(100),
	}
	err = mgo.CreateDocument(index, key, object)
	if err != nil {
		t.Error(err.Error())
		return
	}
	body, err := mgo.QueryDocument(index, key, map[string]interface{}{"id":object.Id})
	if err != nil {
		t.Error(err.Error())
		return
	}
	t.Log(body)
	err = mgo.UpdateDocument(index, key, map[string]interface{}{"id":object.Id}, map[string]interface{}{"Name": "111111", "Age": 100})
	if err != nil {
		t.Error(err.Error())
		return
	}
	err = mgo.DeleteDocument(index, key, map[string]interface{}{"id":object.Id})
	if err != nil {
		t.Error(err.Error())
		return
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
		err = mgo.CreateDocument(index, key, object)
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


