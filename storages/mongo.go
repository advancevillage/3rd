//author: richard
package storages

import (
	"context"
	"encoding/json"
	"github.com/advancevillage/3rd/logs"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"strings"
	"time"
)

const (
	poolSize = 30
	MongoDBTimeout = 30 // 's
	identification = "_3rd_internal_id_"
)

func NewMongoDB(url string, logger logs.Logs) (*MongoDB, error) {
	var err error
	mgo := MongoDB{}
	mgo.logger = logger
	mgo.url = url
	mgo.conn, err = mongo.NewClient(options.Client().ApplyURI(mgo.url).SetMaxPoolSize(poolSize))
	ctx, cancel := context.WithTimeout(context.Background(), MongoDBTimeout * time.Second)
	defer cancel()
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

func (s *MongoDB) CreateStorageV2Exd(index string, key string, field string, body []byte) error {
	var object = make(map[string]interface{})
	var err error
	err = json.Unmarshal(body, &object)
	if err != nil {
		s.logger.Error(err.Error())
		return err
	}
	object[identification] = field
	return s.CreateDocument(index, key, object)
}

func (s *MongoDB) DeleteStorageV2Exd(index string, key string, field ...string) error {
	var where = make(map[string]interface{})
	for i := range key {
		where[identification] = field[i]
		err := s.DeleteDocument(index, key, where)
		if err != nil {
			s.logger.Error(err.Error())
		} else {
			continue
		}
	}
	return nil
}

func (s *MongoDB) UpdateStorageV2Exd(index string, key string, field string, body []byte) error {
	var set = make(map[string]interface{})
	var err error
	err = json.Unmarshal(body, &set)
	if err != nil {
		s.logger.Error(err.Error())
		return err
	}
	var where = make(map[string]interface{})
	where[identification] = field
	return s.UpdateDocument(index, key, where, set)
}

func (s *MongoDB) QueryStorageV2Exd(index string, key string, field  string) ([]byte, error) {
	var where = make(map[string]interface{})
	where[identification] = field
	return s.QueryDocument(index, key, where)
}

func (s *MongoDB) QueryStorageV3(index string, where map[string]interface{}, limit int, offset int, sort map[string]interface{}) ([][]byte, int64, error) {
	return s.QueryDocuments(index, index, where, int64(limit), int64(offset), sort)
}

func (s *MongoDB) CreateDocument(database string, collect string, body interface{}) error {
	buf, err := bson.Marshal(body)
	if err != nil {
		s.logger.Error(err.Error())
		return err
	}
	collection := s.conn.Database(database).Collection(collect)
	ctx, cancel := context.WithTimeout(context.Background(), MongoDBTimeout * time.Second)
	defer cancel()
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
			Key: k,
			Value: where[k],
		}
		d = append(d, e)
	}
	collection := s.conn.Database(database).Collection(collect)
	ctx, cancel := context.WithTimeout(context.Background(), MongoDBTimeout * time.Second)
	defer cancel()
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

//note:
// where[sort] = map[string]string
func (s *MongoDB) QueryDocuments(database string, collect string,  where map[string]interface{}, limit int64, offset int64, sort map[string]interface{}) ([][]byte, int64, error) {
	var d,o  = make(bson.D,0), make(bson.D,0)
	for k :=range where {
		e := bson.E{
			Key: k,
			Value: where[k],
		}
		d = append(d, e)
	}
	for k := range sort {
		e := bson.E{
			Key: k,
			Value: sort[k],
		}
		o = append(o, e)
	}
	option := options.Find()
	option.SetLimit(limit)
	option.SetSkip(offset)
	option.SetSort(o)

	collection := s.conn.Database(database).Collection(collect)
	ctx, cancel := context.WithTimeout(context.Background(), MongoDBTimeout * time.Second)
	defer cancel()
	total, err := collection.CountDocuments(ctx,d)
	if err != nil {
		return nil, 0, err
	}
	cursor, err := collection.Find(ctx, d, option)
	if err != nil {
		return nil, 0, err
	}
	var body = make([][]byte, 0)
	for cursor.Next(ctx) {
		o  := make(map[string]interface{})
		err = cursor.Decode(&o)
		if err != nil {
			s.logger.Error(err.Error())
			continue
		}
		buf, err := json.Marshal(o)
		if err != nil {
			s.logger.Error(err.Error())
			continue
		}
		body = append(body, buf)
	}
	return body, total, nil
}

func (s *MongoDB) UpdateDocument(database string, collect string, where map[string]interface{}, set map[string]interface{}) error {
	var d  bson.D
	for k :=range where {
		e := bson.E{
			Key: k,
			Value: where[k],
		}
		d = append(d, e)
	}
	var ds  bson.D
	for k :=range set {
		e := bson.E{
			Key: k,
			Value: set[k],
		}
		ds = append(ds, e)
	}
	update := bson.D{
		{"$set", ds},
	}
	collection := s.conn.Database(database).Collection(collect)
	ctx, cancel := context.WithTimeout(context.Background(), MongoDBTimeout * time.Second)
	defer cancel()
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
	ctx, cancel := context.WithTimeout(context.Background(), MongoDBTimeout * time.Second)
	defer cancel()
	_, err := collection.DeleteMany(ctx, d)
	if err != nil {
		return err
	}
	return nil
}

