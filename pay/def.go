//author: richard
package pay

import (
	"github.com/advancevillage/3rd/logs"
	"github.com/braintree-go/braintree-go"
)

type IPay interface {

}


type BrainTreePay struct {
	bt  *braintree.Braintree
	logger logs.Logs
}
