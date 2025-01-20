package dbx

import (
	"context"

	dbproxy "github.com/advancevillage/3rd/proto"
)

var _ IDBProxy = (*pg)(nil)

type pg struct {
}

func (p *pg) ExecSql(ctx context.Context, schema string, sqlStr string) ([]*dbproxy.SqlRow, error) {
	return nil, nil
}
