//author: richard
package storages

import (
	"fmt"
	"github.com/advancevillage/3rd/logs"
	"github.com/syndtr/goleveldb/leveldb"
)

func NewLevelDB(schema string, logger logs.Logs) (Storage, error) {
	var err error
	l := &LevelDB{}
	l.logger = logger
	l.schema = schema
	l.conn, err = leveldb.OpenFile(l.schema, nil)
	if err != nil {
		l.logger.Emergency(err.Error())
		return nil, err
	}
	return l, nil
}

//实现接口
func (l *LevelDB) CreateStorage(key string, body []byte) error {
	err := l.conn.Put([]byte(key), body, nil)
	if err != nil {
		l.logger.Error(err.Error())
		return err
	}
	return nil
}

func (l *LevelDB) UpdateStorage(key string, body []byte) error {
	err := l.conn.Put([]byte(key), body, nil)
	if err != nil {
		l.logger.Error(err.Error())
		return err
	}
	return nil
}

func (l *LevelDB) QueryStorage(key  string) ([]byte, error) {
	value, err := l.conn.Get([]byte(key), nil)
	if err != nil {
		l.logger.Error(err.Error())
		return nil, err
	}
	return value, nil
}

func (l *LevelDB) DeleteStorage(key ...string) error {
	var err error
	for i := 0; i < len(key); i++ {
		err = l.conn.Delete([]byte(key[i]), nil)
		if err != nil {
			l.logger.Error(err.Error())
		} else {
			continue
		}
	}
	return nil
}

func (l *LevelDB) CreateStorageV2(index string, key string, body []byte) error {
	return l.CreateStorage(fmt.Sprintf("%s/%s", index, key), body)
}

func (l *LevelDB) UpdateStorageV2(index string, key string, body []byte) error {
	return l.UpdateStorage(fmt.Sprintf("%s/%s", index, key), body)
}

func (l *LevelDB) QueryStorageV2(index string, key  string) ([]byte, error) {
	return l.QueryStorage(fmt.Sprintf("%s/%s", index, key))
}

func (l *LevelDB) DeleteStorageV2(index string, key ...string) error {
	var keys = make([]string, 0, len(key))
	for i := 0; i < len(key); i++ {
		keys = append(keys, fmt.Sprintf("%s/%s", index, key[i]))
	}
	return l.DeleteStorage(keys ...)
}

//TODO
func (l *LevelDB) QueryStorageV3(index string, where map[string]interface{}, limit int, offset int, sort map[string]interface{}) ([][]byte, int64, error) {
	return nil, 0, nil
}