//author: richard
package dbx

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	dbproxy "github.com/advancevillage/3rd/proto"
	_ "github.com/go-sql-driver/mysql"
	"github.com/xwb1989/sqlparser"
)

type IDBProxy interface {
	ExecSql(ctx context.Context, schema string, sqlStr string) ([]*dbproxy.SqlRow, error)
}

//@overview: 基于mysql数据库实现.需要支持的功能断开重连,读写分离. 断开重连功能标准库已经实现.
//@author: richard.sun
//@date: 2020-11
//@param:
// f   数据库连接器
// dns 数据库地址
// db  数据库连接
type mysql struct {
	f   func(addr string) (*sql.DB, error)
	dns string
	db  *sql.DB
}

func connect(dsn string) (*sql.DB, error) {
	var conn *sql.DB
	var err error
	conn, err = sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	conn.SetConnMaxLifetime(time.Minute * 3)
	conn.SetMaxOpenConns(10)
	conn.SetMaxIdleConns(10)

	return conn, nil
}

func query(ctx context.Context, db *sql.DB, sqlStr string) (*dbproxy.ExecuteSqlResponse, error) {
	var (
		err     error
		rows    *sql.Rows
		columns []string
		result  = new(dbproxy.ExecuteSqlResponse)
	)
	//1. 查询结果行数
	rows, err = db.QueryContext(ctx, sqlStr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	//2. 查询结果列数
	columns, err = rows.Columns()
	if err != nil {
		return nil, err
	}
	//3. 解析结果集
	for rows.Next() {
		row := make([][]byte, len(columns))
		dest := make([]interface{}, len(columns))
		for i := range row {
			dest[i] = &row[i]
		}
		err = rows.Scan(dest...)
		if err != nil {
			return nil, err
		}
		result.Rows = append(result.Rows, &dbproxy.SqlRow{Column: row})
	}
	//4. 返回结果集
	return result, nil

}

func exec(ctx context.Context, db *sql.DB, sqlStr string) (*dbproxy.ExecuteSqlResponse, error) {
	var (
		err    error
		affect sql.Result
		result = new(dbproxy.ExecuteSqlResponse)
	)
	//1. 执行SQL
	affect, err = db.ExecContext(ctx, sqlStr)
	if err != nil {
		return nil, err
	}
	//2. LastInsertId
	result.InsertId, err = affect.LastInsertId()
	if err != nil {
		return nil, err
	}
	//3. 影响数据行数
	result.AffectedRows, err = affect.RowsAffected()
	if err != nil {
		return nil, err
	}
	//4. 返回结果
	return result, nil
}

func NewDBProxy(dns string) (IDBProxy, error) {
	var p = &mysql{}
	var err error
	p.f = connect
	p.dns = dns
	p.db, err = p.f(p.dns)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (c *mysql) ExecSql(ctx context.Context, schema string, sqlStr string) ([]*dbproxy.SqlRow, error) {
	var (
		err    error
		stmt   sqlparser.Statement
		result *dbproxy.ExecuteSqlResponse
	)
	//1. 解析SQL
	stmt, err = sqlparser.Parse(sqlStr)
	if err != nil {
		return nil, fmt.Errorf("unrecognized sql statement. %s", sqlStr)
	}
	//2. 读写分开
	switch stmt.(type) {
	case *sqlparser.Select:
		result, err = query(ctx, c.db, sqlStr)
	default:
		result, err = exec(ctx, c.db, sqlStr)
	}
	//3. 错误处理
	if err != nil {
		return nil, err
	}
	//4. 结果返回
	return result.GetRows(), nil
}
