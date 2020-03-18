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
	token, err := brainTreePay.CreateClientToken()
	if err != nil {
		t.Error(err.Error())
	}else {
		t.Log(token)
	}
	err = brainTreePay.Transaction("111")
	if err != nil {
		t.Error(err.Error())
	}
}
