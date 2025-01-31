// author: richard
package dbx_test

import (
	"context"
	"database/sql"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/advancevillage/3rd/dbx"
	"github.com/advancevillage/3rd/logx"
	"github.com/advancevillage/3rd/mathx"
	"github.com/romanyx/polluter"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

//@overview: 构建数据库单元测试的结构
//	create database if not exists  test default charset utf8;
//	create user 'test'@'%' identified by 'password';
//	grant  all on test.* to 'test'@'%';
//	flush privileges;
//
//	use test;
//
//	create table if not exists t_test_user (
//	    id int auto_increment,
//	    name char(64) not null default '' comment '名称',
//	    age  int not null default 0 comment '年龄',
//	    createTime int(64) not null default 0 comment '创建时间',
//	    updateTime int(64) not null default 0 comment '更新时间',
//	    deleteTime int(64) not null default 0 comment '删除时间',
//	    primary key (`id`)
//	)engine=innodb auto_increment=0 default charset utf8;
//
//
//	create table if not exists t_test_class (
//	    cid int auto_increment,
//	    uid int not null default 0 comment '用户id',
//	    name char(64) not null default '' comment '名称',
//	    teacher  char(64) not null default '' comment '教师',
//	    createTime int(64) not null default 0 comment '创建时间',
//	    updateTime int(64) not null default 0 comment '更新时间',
//	    deleteTime int(64) not null default 0 comment '删除时间',
//	    primary key (`cid`)
//	)engine=innodb auto_increment=0 default charset utf8;

type user struct {
	Id         int    `yaml:"id"`
	Name       string `yaml:"name"`
	Age        int    `yaml:"age"`
	CreateTime int64  `yaml:"createTime"`
	UpdateTime int64  `yaml:"updateTime"`
	DeleteTime int64  `yaml:"deleteTime"`
}

type class struct {
	CId        int    `yaml:"cid"`
	UId        int    `yaml:"uid"`
	Name       string `yaml:"name"`
	Teacher    string `yaml:"teacher"`
	CreateTime int64  `yaml:"createTime"`
	UpdateTime int64  `yaml:"updateTime"`
	DeleteTime int64  `yaml:"deleteTime"`
}

var testData = map[string]struct {
	dsn     string
	data    map[string]interface{}
	slt     string
	sltArgs []any
	sltExpt interface{}
	upt     string
	uptArgs []any
	uptExpt interface{}
}{
	"case1": { //单表查询
		dsn: "mysql://test:password@tcp(127.0.0.1:3306)/test",
		data: map[string]interface{}{
			"t_test_user": []user{
				{1, "T1", 11, time.Now().Unix(), time.Now().Unix(), 0},
				{2, "T2", 12, time.Now().Unix(), time.Now().Unix(), 0},
				{3, "T3", 13, time.Now().Unix(), time.Now().Unix(), time.Now().Unix()},
			},
		},
		slt:     `select name, age from t_test_user where deleteTime = ? order by id;`,
		sltArgs: []any{0},
		sltExpt: []interface{}{
			&user{Name: "T1", Age: 11},
			&user{Name: "T2", Age: 12},
		},
		upt:     `update t_test_user set name = ? where id = ? limit 2;`,
		uptArgs: []any{"T4", 1},
		uptExpt: []interface{}{
			&user{Name: "T4", Age: 11},
			&user{Name: "T2", Age: 12},
		},
	},
}

func TestMariaSqlExecutor_ExecSql(t *testing.T) {
	logger, err := logx.NewLogger("debug")
	assert.Nil(t, err)

	ctx := context.WithValue(context.TODO(), logx.TraceId, mathx.UUID())
	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(time.Second*2))
	defer cancel()

	for n, p := range testData {
		f := func(t *testing.T) {
			executor, err := dbx.NewMariaSqlExecutor(ctx, logger, dbx.WithSqlDsn(p.dsn))
			assert.Nil(t, err)

			// 1. 查询
			rows, err := executor.ExecSql(ctx, p.slt, p.sltArgs...)
			assert.Nil(t, err)

			var actual = make([]interface{}, 0, rows.AffectedRows)
			for _, row := range rows.Rows {
				var u = &user{}
				u.Name = string(row.GetColumn()[0])
				u.Age, err = strconv.Atoi(string(row.GetColumn()[1]))
				assert.Nil(t, err)
				actual = append(actual, u)
			}
			assert.Equal(t, p.sltExpt, actual)

			// 2. 更新
			rows, err = executor.ExecSql(ctx, p.upt, p.uptArgs...)
			assert.Nil(t, err)
			assert.Equal(t, int64(1), rows.GetAffectedRows())

			// 3. 查询
			rows, err = executor.ExecSql(ctx, p.slt, p.sltArgs...)
			assert.Nil(t, err)

			actual = make([]interface{}, 0, rows.AffectedRows)
			for _, row := range rows.Rows {
				var u = &user{}
				u.Name = string(row.GetColumn()[0])
				u.Age, err = strconv.Atoi(string(row.GetColumn()[1]))
				assert.Nil(t, err)
				actual = append(actual, u)
			}
			assert.Equal(t, p.uptExpt, actual)

		}
		t.Run(n, func(t *testing.T) {
			prepare(p.data, p.dsn, t, f)
		})
	}
}

func prepare(td map[string]interface{}, dsn string, t *testing.T, f func(*testing.T)) {
	//0. 预先在数据库中创建表
	var tables = []string{"t_test_user", "t_test_class"}
	var result = make([]string, 0, len(td))
	for _, t := range tables {
		data, ok := td[t]
		if ok {
			switch t {
			case "t_test_user":
				rows := data.([]user)
				db := struct {
					Data []user `yaml:"t_test_user"`
				}{
					Data: rows,
				}
				b, err := yaml.Marshal(db)
				if err != nil {
					panic(err)
				}
				result = append(result, string(b))
			case "t_test_class":
				rows := data.([]class)
				db := struct {
					Data []class `yaml:"t_test_class"`
				}{
					Data: rows,
				}
				b, err := yaml.Marshal(db)
				if err != nil {
					panic(err)
				}
				result = append(result, string(b))
			}
		}
	}
	var db, err = sql.Open("mysql", strings.TrimPrefix(dsn, "mysql://"))
	assert.Nil(t, err)
	//2. 构建数据
	var p = polluter.New(polluter.MySQLEngine(db))
	var seed = strings.NewReader(strings.Join(result, ""))
	//3. 数据库注入测试数据
	err = p.Pollute(seed)
	assert.Nil(t, err)
	defer db.Close()
	f(t)
	db.Exec("truncate table t_test_user;")
	db.Exec("truncate table t_test_class;")

}
