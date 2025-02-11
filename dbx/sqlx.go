// author: richard
package dbx

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/advancevillage/3rd/logx"
	_ "github.com/go-sql-driver/mysql"
)

type SqlExecutor interface {
	ExecSql(ctx context.Context, query string, args ...any) (*SqlReply, error)
}

type SqlOption interface {
	apply(*sqlOption)
}

func WithSqlDsn(dsn string) SqlOption {
	return newFuncSqlOption(func(o *sqlOption) {
		o.dsn = dsn
	})
}

type sqlOption struct {
	dsn             string // 组件地址
	connMaxLiftTime int    // 连接最大生命周期
	maxOpenConns    int    // 最大打开连接数
	maxIdleConns    int    // 最大空闲连接数
}

var defaultSqlOptions = sqlOption{
	dsn:             "mysql://test:password@tcp(127.0.0.1:3306)/test",
	connMaxLiftTime: 180, // 秒
	maxOpenConns:    10,
	maxIdleConns:    3,
}

type funcSqlOption struct {
	f func(*sqlOption)
}

func (fdo *funcSqlOption) apply(do *sqlOption) {
	fdo.f(do)
}

func newFuncSqlOption(f func(*sqlOption)) *funcSqlOption {
	return &funcSqlOption{
		f: f,
	}
}

var _ SqlExecutor = (*maria)(nil)

type maria struct {
	conn   *sql.DB
	opts   sqlOption
	logger logx.ILogger
}

func NewMariaSqlExecutor(ctx context.Context, logger logx.ILogger, opt ...SqlOption) (SqlExecutor, error) {
	return newMariaSqlExecutor(ctx, logger, opt...)
}

func newMariaSqlExecutor(ctx context.Context, logger logx.ILogger, opt ...SqlOption) (*maria, error) {
	// 1. 解析参数
	opts := defaultSqlOptions
	for _, o := range opt {
		o.apply(&opts)
	}

	// 2. 连接数据库
	conn, err := sql.Open("mysql", strings.TrimPrefix(opts.dsn, "mysql://"))
	if err != nil {
		logger.Errorw(ctx, "connect sql maria failed", "err", err, "dsn", opts.dsn)
		return nil, err
	}

	// 3. 连接设置
	conn.SetConnMaxLifetime(time.Second * time.Duration(opts.connMaxLiftTime))
	conn.SetMaxOpenConns(opts.maxOpenConns)
	conn.SetMaxIdleConns(opts.maxIdleConns)

	// 4. 测试连接
	err = conn.Ping()
	if err != nil {
		logger.Errorw(ctx, "ping sql maria failed", "err", err, "dsn", opts.dsn)
		return nil, err
	}

	// 5. 返回对象
	return &maria{
		conn:   conn,
		opts:   opts,
		logger: logger,
	}, nil
}

func (c *maria) ExecSql(ctx context.Context, query string, args ...any) (*SqlReply, error) {
	query = strings.ReplaceAll(query, "\t\n", " ")
	query = strings.TrimSpace(query)
	c.logger.Infow(ctx, "exec sql", "query", query, "args", args)
	switch strings.ToLower(query[:6]) {
	case "select":
		return c.query(ctx, query, args...)

	default:
		// insert update delete
		return c.exec(ctx, query, args...)
	}
}

func (c *maria) query(ctx context.Context, query string, args ...any) (*SqlReply, error) {
	var (
		err     error
		rows    *sql.Rows
		columns []string
		result  = new(SqlReply)
	)
	// 1. 预编译SQL
	stmt, err := c.conn.PrepareContext(ctx, query)
	if err != nil {
		c.logger.Errorw(ctx, "prepare sql maria failed", "err", err, "query", query)
		return nil, err
	}
	defer stmt.Close()

	// 2. 执行SQL
	rows, err = stmt.QueryContext(ctx, args...)
	if err != nil {
		c.logger.Errorw(ctx, "query sql maria failed", "err", err, "query", query)
		return nil, err
	}
	defer rows.Close()

	// 3. 查询结果列数
	columns, err = rows.Columns()
	if err != nil {
		c.logger.Errorw(ctx, "query sql maria failed", "err", err, "query", query)
		return nil, err
	}

	// 4. 解析结果集
	for rows.Next() {
		row := make([]string, len(columns))
		dest := make([]any, len(columns))
		for i := range row {
			dest[i] = &row[i]
		}
		err = rows.Scan(dest...)
		if err != nil {
			return nil, err
		}
		result.Rows = append(result.Rows, &SqlRow{Column: row})
	}

	// 5. 返回结果集
	result.AffectedRows = int64(len(result.Rows))
	return result, nil
}

func (c *maria) exec(ctx context.Context, query string, args ...any) (*SqlReply, error) {
	var (
		err    error
		affect sql.Result
		result = new(SqlReply)
	)
	// 1. 预编译SQL
	stmt, err := c.conn.PrepareContext(ctx, query)
	if err != nil {
		c.logger.Errorw(ctx, "prepare sql maria failed", "err", err, "query", query)
		return nil, err
	}
	defer stmt.Close()

	// 2. 执行SQL
	affect, err = stmt.ExecContext(ctx, args...)
	if err != nil {
		c.logger.Errorw(ctx, "exec sql maria failed", "err", err, "query", query)
		return nil, err
	}

	// 3. LastInsertId
	result.InsertId, err = affect.LastInsertId()
	if err != nil {
		c.logger.Errorw(ctx, "exec sql maria failed", "err", err)
		return nil, err
	}

	// 4. 影响数据行数
	result.AffectedRows, err = affect.RowsAffected()
	if err != nil {
		c.logger.Errorw(ctx, "exec sql maria failed", "err", err)
		return nil, err
	}

	// 5. 返回结果
	return result, nil
}
