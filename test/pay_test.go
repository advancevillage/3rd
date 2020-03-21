//author: richard
package test

import (
	"github.com/advancevillage/3rd/logs"
	 "github.com/advancevillage/3rd/pay"
	"github.com/advancevillage/3rd/utils"
	"testing"
)

func TestBrainTreePay(t *testing.T) {
	logger := logs.NewStdLogger()
	//https://payments.sandbox.braintree-api.com/graphql test
	//https://payments.braintree-api.com/graphql production
	url := "https://payments.sandbox.braintree-api.com/graphql"
	brainTreePay := pay.NewBrainTree(url, "34f8k2dmry3hcs9m","mdx8ssqjhvcgpywt", "08a87176192595325cbf076296aa5b53", logger)
	var token string
	err := brainTreePay.ClientToken(&token)
	if err != nil {
		t.Error(err.Error())
	}
	translate := make(map[string]string)
	//付款
	err = brainTreePay.Transaction("tokencc_bh_x2g46s_ktm4q5_zktnjq_j7y4b2_kd7", 20,&translate)
	if err != nil {
		t.Error(err.Error())
	}
	//查询交易流水状态
	err = brainTreePay.TransactionStatus(translate["id"], &translate)
	if err != nil {
		t.Error(err.Error())
	}
	//退款
	callback := make(map[string]string)
	refundOrderId := utils.SnowFlakeIdString()
	err = brainTreePay.Refund("dHJhbnNhY3Rpb25fZ2hoejBrZHg", 1, refundOrderId, &callback)
	if err != nil {
		t.Error(err.Error())
	}
}
