// author: richard
package dbx

import (
	"context"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

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

type testDBProxy struct {
	dbCli IDBProxy
}

func newTestDBProxyService() (*testDBProxy, error) {
	var s = testDBProxy{}
	return &s, nil
}

// @overview: 单元测试数据库接口
// @param:
//
//	td(test data) 测试数据. yaml格式
func (s *testDBProxy) RunInDBProxy(td map[string]interface{}, t *testing.T, f func(*testing.T)) {
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
	//1. 将map[string]interface{} 转换成 yaml 字符串
	var seed = strings.Join(result, "")
	//2. 调用函数
	RunUnitTest(seed, t, func(t *testing.T, dbCli IDBProxy) {
		s.dbCli = dbCli
		f(t)
	})
}

// @overview: 单元测试Case
// case1 单表查询 case2 多表查询 case3 单表更新 case4 多表更新 case5 单表删除 case6 多表删除
var testData = map[string]struct {
	data     map[string]interface{}
	sqlStr   string
	excepted interface{}
}{
	"case1": { //单表查询
		data: map[string]interface{}{
			"t_test_user": []user{
				{1, "T1", 11, time.Now().Unix(), time.Now().Unix(), 0},
				{2, "T2", 12, time.Now().Unix(), time.Now().Unix(), 0},
				{3, "T3", 13, time.Now().Unix(), time.Now().Unix(), time.Now().Unix()},
			},
		},
		sqlStr: `
			select name, age
			from t_test_user
			where deleteTime = 0
			order by id;
		`,
		excepted: []interface{}{
			&user{Name: "T1", Age: 11},
			&user{Name: "T2", Age: 12},
		},
	},
}

func TestMysql_ExecSql(t *testing.T) {
	var s, err = newTestDBProxyService()
	if err != nil {
		t.Fatal(err)
		return
	}
	for n, p := range testData {
		ut := func(t *testing.T) {
			var ctx, cancel = context.WithTimeout(context.TODO(), 3*time.Second)
			defer cancel()

			var rows, err = s.dbCli.ExecSql(ctx, testSchema, p.sqlStr)
			if err != nil {
				t.Fatal(err)
				return
			}

			var actual = make([]interface{}, 0, len(rows))

			for _, row := range rows {
				var u = &user{}
				u.Name = string(row.GetColumn()[0])
				u.Age, err = strconv.Atoi(string(row.GetColumn()[1]))
				if err != nil {
					t.Fatal(err)
					return
				}
				actual = append(actual, u)
			}

			assert.Equal(t, p.excepted, actual)
		}
		t.Run(n, func(t *testing.T) {
			s.RunInDBProxy(p.data, t, ut)
		})
	}
}
