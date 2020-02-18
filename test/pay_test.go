//author: richard
package test

import (
	"github.com/advancevillage/3rd/logs"
	 "github.com/advancevillage/3rd/pay"
	"testing"
)

func TestBrainTreePay(t *testing.T) {
	logger := logs.NewStdLogger()
	brainTreePay := pay.NewBrainTree("34f8k2dmry3hcs9m", "mdx8ssqjhvcgpywt", "08a87176192595325cbf076296aa5b53", logger)
	brainTreePay.Transaction()
}
