//author: richard
package test

import (
	"3rd/logs"
	"3rd/rpcs"
	"testing"
)

func TestRpcServer_StartServer(t *testing.T) {
	logger, err := logs.NewTxtLogger("rpc.log", 64, 3)
	if err != nil {
		t.Error(err.Error())
	}
	rcvr := make([]interface{}, 0)
	rcvr = append(rcvr, new(Edwin))
	server := rpcs.NewServer("localhost", 11311, logger, rcvr)
	err = server.StartServer()
	if err != nil {
		t.Error(err.Error())
	}
}


type Edwin int
// 定义 rpc method 第一个参数是请求对象，
// 第二参数是返回对象, 返回值是返回 rpc
// 内部调用过程中出现的错误信息
func (this *Edwin) Add(args map[string]float64,res *float64) error  {
   *res = args["num1"] + args["num2"]
   return  nil
}
func (this *Edwin) Multi(args map[string]interface{},res *float64) error {
	*res = args["num1"].(float64) * args["num2"].(float64)
	return nil
}