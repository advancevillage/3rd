//author: richard
package storages

import (
	"3rd/logs"
	"3rd/utils"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

func NewMysql(master *Database, slaves []*Database, logger logs.Logs) (*Mysql, error) {
	var err error
	mysql := &Mysql{}
	mysql.master = master
	mysql.slaves = slaves
	mysql.logger = logger
	//初始化Mysql连接
	//eg: username:password@protocol(address)/dbname?param=value
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s", master.Username, master.Password, master.Host, master.Port, master.Schema, master.Charset)
	master.conn, err = sql.Open(master.Driver, dsn)
	if err != nil {
		mysql.logger.Emergency(err.Error())
		return nil, err
	}
	for i := 0; i < len(slaves); i++ {
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s", slaves[i].Username, slaves[i].Password, slaves[i].Host, slaves[i].Port, slaves[i].Schema, slaves[i].Charset)
		slaves[i].conn, err = sql.Open(master.Driver, dsn)
		if err != nil {
			mysql.logger.Emergency(err.Error())
			return nil, err
		}
	}
	return mysql, nil
}

//readonly: flag = false
//write&read: flag = true
func (m *Mysql) Connection(flag bool) *sql.DB {
	if flag {
		return  m.master.conn
	} else {
		index := utils.RandsInt(len(m.slaves))
		return m.slaves[index].conn
	}
}
