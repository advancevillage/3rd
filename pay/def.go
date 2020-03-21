//author: richard
package pay

import (
	"github.com/advancevillage/3rd/https"
	"github.com/advancevillage/3rd/logs"
)

type IPay interface {
	ClientToken(callback *string) error
	Transaction(nonce string, amount float64, callback *map[string]string) error
	TransactionStatus(transactionId string, callback *map[string]string) error
	Refund(transactionId string, amount float64, refundOrderId string, callback *map[string]string) error
}


type BrainTreePay struct {
	client *https.Client
	merchant string
	urlRoot string
	logger  logs.Logs
}

type brainTreeResponse struct {
	Data       map[string]interface{}   `json:"data"`
	Errors     []map[string]interface{} `json:"errors"`
	Extensions  map[string]interface{}  `json:"extensions"`
}
