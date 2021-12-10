//author: richard
package dbx

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-txdb"
	"github.com/romanyx/polluter"
)

var (
	testDB = "mysql://test:password@tcp(localhost:3306)/test"
)

const testSchema = "test"

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

func init() {
	if os.Getenv("TEST_DB") != "" {
		testDB = os.Getenv("TEST_DB")
	}
	txdb.Register(testSchema, "mysql", strings.TrimPrefix(testDB, "mysql://"))
}

type DBProxyTestFunc func(t *testing.T, dbCli IDBProxy)

//@overview: Mainly this package was created for testing purposes, to give the ability
//to seed a database with records from simple .yaml files.
//@author: richard.sun
//@date: 2020-11
//@param:
// yaml	string yaml 格式
//      const input = `
//      roles:
//      - name: User
//      users:
//      - name: Roman
//        role_id: 1
//      `
func RunUnitTest(yaml string, t *testing.T, f DBProxyTestFunc) {
	if testing.Short() {
		t.SkipNow()
	}
	//1. 建立测试连接
	var schema = fmt.Sprintf("connection_%d", time.Now().UnixNano())
	var db, err = sql.Open(testSchema, schema)
	if err != nil {
		t.Fatal(err)
		return
	}
	//2. 构建数据
	var p = polluter.New(polluter.MySQLEngine(db))
	var seed = strings.NewReader(yaml)
	//3. 数据库注入测试数据
	if err := p.Pollute(seed); err != nil {
		t.Fatalf("failed to pollute: %s", err)
		return
	}
	defer db.Close()
	//4. 回调函数
	var dbCli = &mysql{db: db}
	f(t, dbCli)
}
