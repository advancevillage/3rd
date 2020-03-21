//author: richard
package pay

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/advancevillage/3rd/https"
	"github.com/advancevillage/3rd/logs"
)

func NewBrainTree(url string, merchant string, public string, private string, logger logs.Logs) IPay {
	header := make(map[string]string)
	header["Braintree-Version"] = "2019-01-01"
	header["Content-Type"]      = "application/json"
	header["Authorization"] = "Basic " + base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", public, private)))
	return &BrainTreePay{
		logger: logger,
		urlRoot: url,
		merchant: merchant,
		client: https.NewRequest(header, 10, 3),
	}
}

func (s *BrainTreePay) ClientToken(callback *string) error {
	params := make(map[string]interface{})
	params["query"] = "" +
		"mutation CreateClientToken($input: CreateClientTokenInput){" +
		"	createClientToken(input: $input){" +
		"		clientToken" +
		"	}" +
		"}"
	params["variables"] = map[string]interface{} {
		"input": map[string]interface{} {
			"clientToken": map[string]interface{} {
				"merchantAccountId": s.merchant,
			},
		},
	}
	buf, err := json.Marshal(params)
	if err != nil {
		return err
	}
	body, err := s.client.POST(s.urlRoot, nil, buf)
	if err != nil {
		s.logger.Alert(err.Error())
		return err
	}
	brain, err := s.response(body)
	if err != nil {
		s.logger.Alert(err.Error())
		return err
	}
	if v, ok := brain.Extensions["requestId"]; ok {
		s.logger.Info("BrainTree requestId: ", v)
	}
	for i := range brain.Errors {
		buf, err := json.Marshal(brain.Errors[i])
		if err != nil {
			continue
		} else {
			s.logger.Alert("BrainTree Error: %s", string(buf))
		}
	}
	if len(brain.Errors) > 0 {
		return errors.New("response data error")
	}
	i, ok := brain.Data["createClientToken"]
	if !ok {
		s.logger.Error("BrainTree Error: %s", "response data error")
		return errors.New("response data error")
	}
	buf, err = json.Marshal(i)
	if err != nil {
		return err
	}
	v := make(map[string]string)
	err = json.Unmarshal(buf, &v)
	if err != nil {
		return err
	}
	if _, ok := v["clientToken"]; !ok {
		return errors.New("response data error")
	}
	*callback = v["clientToken"]
	return nil
}

//美元结算
func (s *BrainTreePay) Transaction(nonce string, amount float64, callback *map[string]string) error {
	params := make(map[string]interface{})
	params["query"] = "" +
		"mutation Payment($input: ChargePaymentMethodInput!) {" +
		"	chargePaymentMethod(input: $input) {" +
		"		transaction {" +
		"			id " +
		"  			status " +
		"		}" +
		"   }" +
		"}"
	params["variables"] = map[string]interface{} {
		"input": map[string]interface{} {
			"paymentMethodId": nonce,
			"transaction": map[string]interface{} {
				"amount": amount,
			},
		},
	}
	buf, err := json.Marshal(params)
	if err != nil {
		return err
	}
	body, err := s.client.POST(s.urlRoot, nil, buf)
	if err != nil {
		s.logger.Alert(err.Error())
		return err
	}
	brain, err := s.response(body)
	if err != nil {
		s.logger.Alert(err.Error())
		return err
	}
	if v, ok := brain.Extensions["requestId"]; ok {
		s.logger.Info("BrainTree requestId: ", v)
	}
	for i := range brain.Errors {
		buf, err := json.Marshal(brain.Errors[i])
		if err != nil {
			continue
		} else {
			s.logger.Alert("BrainTree Error: %s", string(buf))
		}
	}
	if len(brain.Errors) > 0 {
		return errors.New("response data error")
	}
	i, ok := brain.Data["chargePaymentMethod"]
	if !ok {
		s.logger.Error("BrainTree Error: %s", "response data error")
		return errors.New("response data error")
	}
	buf, err = json.Marshal(i)
	if err != nil {
		return err
	}
	j := make(map[string]interface{})
	err = json.Unmarshal(buf, &j)
	if err != nil {
		return err
	}
	k, ok := j["transaction"]
	if !ok {
		return errors.New("response data error")
	}
	buf, err = json.Marshal(k)
	if err != nil {
		return err
	}
	v := make(map[string]string)
	err = json.Unmarshal(buf, &v)
	if err != nil {
		return err
	}
	if _, ok := v["id"]; !ok {
		return errors.New("response data error")
	}
	if _, ok := v["status"]; !ok {
		return errors.New("response data error")
	}
	*callback = v
	return nil
}

func (s *BrainTreePay) TransactionStatus(transactionId string, callback *map[string]string) error {
	params := make(map[string]interface{})
	params["query"] = "{" +
		"node(id: "+ fmt.Sprintf("\"%s\"", transactionId) + ") {" +
		"	... on Transaction {" +
		"		id" +
		"		status" +
		"	}" +
		"}" +
		"}"
	buf, err := json.Marshal(params)
	if err != nil {
		return err
	}
	body, err := s.client.POST(s.urlRoot, nil, buf)
	if err != nil {
		s.logger.Alert(err.Error())
		return err
	}
	brain, err := s.response(body)
	if err != nil {
		s.logger.Alert(err.Error())
		return err
	}
	if v, ok := brain.Extensions["requestId"]; ok {
		s.logger.Info("BrainTree requestId: ", v)
	}
	for i := range brain.Errors {
		buf, err := json.Marshal(brain.Errors[i])
		if err != nil {
			continue
		} else {
			s.logger.Alert("BrainTree Error: %s", string(buf))
		}
	}
	if len(brain.Errors) > 0 {
		return errors.New("response data error")
	}
	i, ok := brain.Data["node"]
	if !ok {
		s.logger.Error("BrainTree Error: %s", "response data error")
		return errors.New("response data error")
	}
	buf, err = json.Marshal(i)
	if err != nil {
		return err
	}
	v := make(map[string]string)
	err = json.Unmarshal(buf, &v)
	if err != nil {
		return err
	}
	if _, ok := v["id"]; !ok {
		return errors.New("response data error")
	}
	if _, ok := v["status"]; !ok {
		return errors.New("response data error")
	}
	*callback = v
	return nil
}

//退款结算
func (s *BrainTreePay) Refund(transactionId string, amount float64, refundOrderId string, callback *map[string]string) error {
	params := make(map[string]interface{})
	params["query"] = "" +
		"mutation Refund($input: RefundTransactionInput!){" +
		"	refundTransaction(input: $input){" +
		"		refund {" +
		"			id " +
		"			status " +
		"			orderId " +
		"			amount {" +
		"				value " +
		"			}" +
		"		}" +
		"   }" +
		"}"
	params["variables"] = map[string]interface{} {
		"input": map[string]interface{} {
			"transactionId": transactionId,
			"refund": map[string]interface{} {
				"amount": amount,
				"orderId": refundOrderId,
			},
		},
	}
	buf, err := json.Marshal(params)
	if err != nil {
		return err
	}
	body, err := s.client.POST(s.urlRoot, nil, buf)
	if err != nil {
		s.logger.Alert(err.Error())
		return err
	}
	brain, err := s.response(body)
	if err != nil {
		s.logger.Alert(err.Error())
		return err
	}
	if v, ok := brain.Extensions["requestId"]; ok {
		s.logger.Info("BrainTree requestId: ", v)
	}
	for i := range brain.Errors {
		buf, err := json.Marshal(brain.Errors[i])
		if err != nil {
			continue
		} else {
			s.logger.Alert("BrainTree Error: %s", string(buf))
		}
	}
	if len(brain.Errors) > 0 {
		return errors.New("response data error")
	}
	i, ok := brain.Data["refundTransaction"]
	if !ok {
		s.logger.Error("BrainTree Error: %s", "response data error")
		return errors.New("response data error")
	}
	buf, err = json.Marshal(i)
	if err != nil {
		return err
	}
	j := make(map[string]interface{})
	err = json.Unmarshal(buf, &j)
	if err != nil {
		return err
	}
	k, ok := j["refund"]
	if !ok {
		return errors.New("response data error")
	}
	buf, err = json.Marshal(k)
	if err != nil {
		return err
	}
	v := make(map[string]interface{})
	err = json.Unmarshal(buf, &v)
	if err != nil {
		return err
	}
	if _, ok := v["id"]; !ok {
		return errors.New("response data error")
	}
	if _, ok := v["status"]; !ok {
		return errors.New("response data error")
	}
	if _, ok := v["amount"]; !ok {
		return errors.New("response data error")
	}
	back := make(map[string]string)
	buf, err = json.Marshal(v["amount"])
	if err != nil {
		return err
	}
	err = json.Unmarshal(buf, &back)
	if err != nil {
		return err
	}
	if _, ok := back["value"]; !ok {
		return errors.New("response data error")
	}
	back["amount"] = back["value"]
	back["id"]     = v["id"].(string)
	back["status"] = v["status"].(string)
	delete(back, "value")
	*callback = back
	return nil
}

func (s *BrainTreePay) response(body []byte) (*brainTreeResponse, error) {
	brain := brainTreeResponse{}
	err := json.Unmarshal(body, &brain)
	if err != nil {
		return nil, err
	}
	return &brain, nil
}
