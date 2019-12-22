//author: richard
package storages

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/advancevillage/3rd/logs"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"strings"
	"time"
)

const (
	MongoDBTimeout = 20 // 's
	identification = "_3rd_internal_id_"
)

func NewMongoDB(host string, port int, logger logs.Logs) (*MongoDB, error) {
	var err error
	mgo := MongoDB{}
	mgo.logger = logger
	mgo.host = host
	mgo.port = port
	mgo.conn, err = mongo.NewClient(options.Client().ApplyURI(fmt.Sprintf("mongodb://%s:%d", mgo.host, mgo.port)))
	ctx, _ := context.WithTimeout(context.Background(), MongoDBTimeout * time.Second)
	if err != nil {
		mgo.logger.Error(err.Error())
		return nil, err
	}
	err = mgo.conn.Connect(ctx)
	if err != nil {
		mgo.logger.Error(err.Error())
		return nil, err
	}
	err = mgo.conn.Ping(ctx, readpref.Primary())
	if err != nil {
		mgo.logger.Error(err.Error())
		return nil, err
	}
	return &mgo, nil
}

func (s *MongoDB) CreateStorage(key string, body []byte) error {
	return s.CreateStorageV2(ESDefaultIndex, key, body)
}

func (s *MongoDB) DeleteStorage(key ...string) error {
	return s.DeleteStorageV2(ESDefaultIndex, key ...)
}

func (s *MongoDB) UpdateStorage(key string, body []byte) error {
	return s.UpdateStorageV2(ESDefaultIndex, key, body)
}

func (s *MongoDB) QueryStorage(key  string) ([]byte, error) {
	return s.QueryStorageV2(ESDefaultIndex, key)
}

func (s *MongoDB) CreateStorageV2(index string, key string, body []byte) error {
	var object = make(map[string]interface{})
	var err error
	err = json.Unmarshal(body, &object)
	if err != nil {
		s.logger.Error(err.Error())
		return err
	}
	object[identification] = key
	return s.CreateDocument(index, index, object)
}

func (s *MongoDB) DeleteStorageV2(index string, key ...string) error {
	var where = make(map[string]interface{})
	for i := range key {
		where[identification] = key[i]
		err := s.DeleteDocument(index, index, where)
		if err != nil {
			s.logger.Error(err.Error())
		} else {
			continue
		}
	}
	return nil
}

func (s *MongoDB) UpdateStorageV2(index string, key string, body []byte) error {
	var set = make(map[string]interface{})
	var err error
	err = json.Unmarshal(body, &set)
	if err != nil {
		s.logger.Error(err.Error())
		return err
	}
	var where = make(map[string]interface{})
	where[identification] = key
	return s.UpdateDocument(index, index, where, set)
}

func (s *MongoDB) QueryStorageV2(index string, key  string) ([]byte, error) {
	var where = make(map[string]interface{})
	where[identification] = key
	return s.QueryDocument(index, index, where)
}

func (s *MongoDB) CreateDocument(database string, collect string, body interface{}) error {
	buf, err := bson.Marshal(body)
	if err != nil {
		s.logger.Error(err.Error())
		return err
	}
	collection := s.conn.Database(database).Collection(collect)
	ctx, _ := context.WithTimeout(context.Background(), MongoDBTimeout * time.Second)
	_, err = collection.InsertOne(ctx, buf)
	if err != nil {
		s.logger.Error(err.Error())
		return err
	}
	return nil
}

func (s *MongoDB) QueryDocument(database string, collect string,  where map[string]interface{}) ([]byte, error) {
	var o  = make(map[string]interface{})
	var d  bson.D
	for k :=range where {
		e := bson.E{
			Key: strings.ToLower(k),
			Value: where[k],
		}
		d = append(d, e)
	}
	collection := s.conn.Database(database).Collection(collect)
	ctx, _ := context.WithTimeout(context.Background(), MongoDBTimeout * time.Second)
	err := collection.FindOne(ctx, d).Decode(&o)
	if err != nil {
		return nil, err
	}
	body, err := json.Marshal(o)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (s *MongoDB) UpdateDocument(database string, collect string, where map[string]interface{}, set map[string]interface{}) error {
	var d  bson.D
	for k :=range where {
		e := bson.E{
			Key: strings.ToLower(k),
			Value: where[k],
		}
		d = append(d, e)
	}
	var ds  bson.D
	for k :=range set {
		e := bson.E{
			Key: strings.ToLower(k),
			Value: set[k],
		}
		ds = append(ds, e)
	}
	update := bson.D{
		{"$set", ds},
	}
	collection := s.conn.Database(database).Collection(collect)
	ctx, _ := context.WithTimeout(context.Background(), MongoDBTimeout * time.Second)
	_, err := collection.UpdateOne(ctx, d, update)
	if err != nil {
		return err
	}
	return nil
}

func (s *MongoDB) DeleteDocument(database string, collect string, where map[string]interface{}) error {
	var d  bson.D
	for k :=range where {
		e := bson.E{
			Key: strings.ToLower(k),
			Value: where[k],
		}
		d = append(d, e)
	}
	collection := s.conn.Database(database).Collection(collect)
	ctx, _ := context.WithTimeout(context.Background(), MongoDBTimeout * time.Second)
	_, err := collection.DeleteMany(ctx, d)
	if err != nil {
		return err
	}
	return nil
}

