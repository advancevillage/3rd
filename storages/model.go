//author: richard
package storages

import (
	"3rd/logs"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

type Storage interface {
	prepare(query string) (string, []interface{}, error) //预编译
}

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

type Mysql struct {
	master *Database
	slaves []*Database
	logger logs.Logs
}