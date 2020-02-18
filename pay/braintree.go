//author: richard
package pay

import (
	"context"
	"github.com/advancevillage/3rd/logs"
	"github.com/braintree-go/braintree-go"
	"time"
)

func NewBrainTree(merchantId string, publicKey string, privateKey string, logger logs.Logs) *BrainTreePay {
	bt := braintree.New(
		braintree.Sandbox,
		merchantId,
		publicKey,
		privateKey,)
	return &BrainTreePay{bt: bt, logger: logger}
}

//交易接口
//信用卡 CreditCard
//金额   Money
//返回交易信息  Transaction
func (s *BrainTreePay) Transaction() {
	ctx, cancel := context.WithTimeout(context.Background(), 600 * time.Second)
	defer cancel()
	tx, err := s.bt.Transaction().Create(ctx, &braintree.TransactionRequest{
		Type: "sale",
		Amount: braintree.NewDecimal(99, 2), // 100 cents
		CreditCard: &braintree.CreditCard{
			Number:         "378282246310005",
			ExpirationDate: "05/20",
		},
	})
	if err != nil {
		s.logger.Alert(err.Error())
		return
	}
	s.logger.Info("%s %s %s %s", tx.Id, tx.Channel, tx.Status, tx.Type)
}
