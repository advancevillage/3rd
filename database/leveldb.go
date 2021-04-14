package database

import (
	"fmt"
	"strings"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/storage"
	"github.com/syndtr/goleveldb/leveldb/util"
)

const (
	expiration   = time.Hour * 24
	cleanupCycle = time.Hour

	actionGet = "get"
	actionPut = "put"
	actionDel = "del"
)

type ILevelDB interface {
	Del(key []byte) error
	Put(key []byte, value []byte) error
	Get(key []byte) ([]byte, error)
	Range(prefix []byte) (map[string]string, error)
}

type ldb struct {
	db   *leveldb.DB //并发安全
	quit chan struct{}
}

//@overview: 创建文件存储
//@author: richard.sun
//@param:
//1. dir 文件绝对路径
func NewPersistentStore(dir string) (ILevelDB, error) {
	//1. 参数校验
	if len(dir) <= 0 {
		return nil, fmt.Errorf("dirpath(%s) is invalid", dir)
	}
	var db, err = leveldb.OpenFile(dir, &opt.Options{
		OpenFilesCacheCapacity: 5,
	})
	if _, iscorrupted := err.(*errors.ErrCorrupted); iscorrupted {
		db, err = leveldb.RecoverFile(dir, nil)
	}
	if err != nil {
		return nil, fmt.Errorf("openfile fail. %s", err.Error())
	}
	//2. 创建对象
	var i = &ldb{}
	i.db = db
	i.quit = make(chan struct{})

	return i, nil
}

func NewMemoryStore() (ILevelDB, error) {
	var db, err = leveldb.Open(storage.NewMemStorage(), nil)
	if err != nil {
		return nil, err
	}
	var i = &ldb{}
	i.db = db
	i.quit = make(chan struct{})
	return i, nil
}

func (l *ldb) action(act string, key []byte, value []byte) ([]byte, error) {
	var err error
	switch strings.ToLower(act) {
	case actionGet:
		value, err = l.db.Get(key, nil)
	case actionPut:
		err = l.db.Put(key, value, nil)
	case actionDel:
		err = l.db.Delete(key, nil)
	default:
		err = fmt.Errorf("don't support %s action", act)
	}
	if err != nil {
		return nil, err
	}
	return value, nil
}

func (l *ldb) Put(key []byte, value []byte) error {
	if len(key) <= 0 || len(value) <= 0 {
		return nil
	}
	var _, err = l.action(actionPut, key, value)
	return err
}

func (l *ldb) Get(key []byte) ([]byte, error) {
	if len(key) <= 0 {
		return nil, nil
	}
	return l.action(actionGet, key, nil)
}

func (l *ldb) Del(key []byte) error {
	if len(key) <= 0 {
		return nil
	}
	var _, err = l.action(actionDel, key, nil)
	return err
}

func (l *ldb) Range(prefix []byte) (map[string]string, error) {
	var iter = l.db.NewIterator(util.BytesPrefix(prefix), nil)
	var kv = make(map[string]string)
	for iter.Next() {
		var key = iter.Key()
		var value = iter.Value()
		kv[string(key)] = string(value)
	}
	iter.Release()
	var err = iter.Error()
	if err != nil {
		return nil, err
	}
	return kv, nil
}
