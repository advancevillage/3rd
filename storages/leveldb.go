//author: richard
package storages

import (
	"3rd/logs"
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
