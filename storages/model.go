//author: richard
package storages

import (
	"3rd/logs"
	"database/sql"
	"errors"
	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
	"github.com/olivere/elastic/v7"
	"github.com/syndtr/goleveldb/leveldb"
)

var (
	ErrorKeyNotExist = errors.New("key not exist")
)

const (
	ESDefaultIndex = "20170729"
)

type Database struct {
	Host string
	Port int
	Username string
	Password string
	Schema   string
	Charset  string
	Driver   string
	conn     *sql.DB
}

type Storage interface {
	UpdateStorage(key string, body []byte) error
	CreateStorage(key string, body []byte) error
	QueryStorage(key  string) ([]byte, error)
	DeleteStorage(key ...string) error
}

type Mysql struct {
	master *Database
	slaves []*Database
	logger logs.Logs
}

type Redis struct {
	host   string
	port   int
	auth   string
	schema int
	conn *redis.Client
	logger logs.Logs
}

type LevelDB struct {
	conn *leveldb.DB
	logger logs.Logs
	schema string
}

type TES struct {
	index    string		//default index
	urls   []string
	logger logs.Logs
	conn *elastic.Client
}