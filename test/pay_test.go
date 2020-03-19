//author: richard
package test

import (
	"github.com/advancevillage/3rd/logs"
	 "github.com/advancevillage/3rd/pay"
	"testing"
)

func TestBrainTreePay(t *testing.T) {
	logger := logs.NewStdLogger()
	//https://payments.sandbox.braintree-api.com/graphql test
	//https://payments.braintree-api.com/graphql production
	url := "https://payments.sandbox.braintree-api.com/graphql"
	brainTreePay := pay.NewBrainTree(url, "34f8k2dmry3hcs9m","mdx8ssqjhvcgpywt", "08a87176192595325cbf076296aa5b53", logger)
	var token string
	err := brainTreePay.CreateClientToken(&token)
	if err != nil {
		t.Error(err.Error())
	}
	translate := make(map[string]string)
	//付款
	err = brainTreePay.Payment("tokencc_bj_c9dvxc_248vt8_yst3pg_rydmsj_49z", 5,&translate)
	if err != nil {
		t.Error(err.Error())
	}
	//退款 dHJhbnNhY3Rpb25fZ2EyZWczN2M
	callback := make(map[string]string)
	err = brainTreePay.Refund("dHJhbnNhY3Rpb25fZ2EyZWczN2M", 0.1, &callback)
	if err != nil {
		t.Error(err.Error())
	}
}
