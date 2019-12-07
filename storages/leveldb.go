//author: richard
package storages

import (
	"github.com/advancevillage/3rd/logs"
	"github.com/syndtr/goleveldb/leveldb"
)

func NewLevelDB(schema string, logger logs.Logs) (*LevelDB, error) {
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