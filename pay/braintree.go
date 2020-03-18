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

func NewBrainTree(url string, merchant string, public string, private string, logger logs.Logs) *BrainTreePay {
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

//交易接口
//信用卡 CreditCard
//金额   Money
//返回交易信息  Transaction
func (s *BrainTreePay) CreateClientToken() (string, error) {
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
		return "", err
	}
	body, err := s.client.POST(s.urlRoot, nil, buf)
	if err != nil {
		s.logger.Alert(err.Error())
		return "", err
	}
	brain, err := s.response(body)
	if err != nil {
		s.logger.Alert(err.Error())
		return "", err
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
	i, ok := brain.Data["createClientToken"]
	if !ok {
		s.logger.Error("BrainTree Error: %s", "response data error")
		return "", errors.New("response data error")
	}
	buf, err = json.Marshal(i)
	if err != nil {
		return "", err
	}
	v := make(map[string]string)
	err = json.Unmarshal(buf, &v)
	if err != nil {
		return "", err
	}
	if _, ok := v["clientToken"]; !ok {
		return "", errors.New("response data error")
	}
	return v["clientToken"], nil
}

func (s *BrainTreePay) Transaction(nonce string) error {
	params := make(map[string]interface{})
	params["query"] = "" +
		"mutation VaultWithTypeFragment($input: VaultPaymentMethodInput) {" +
		"	vaultPaymentMethod(input: $input) {" +
		"		paymentMethod {" +
		"			id " +
		"  			usage " +
		"			details {" +
		"				__typename " +
		"			} " +
		"			verification {" +
		"				status" +
		"			}" +
		"		}" +
		"   }" +
		"}"
	params["variables"] = map[string]interface{} {
		"input": map[string]interface{} {
			"paymentMethodId": nonce,
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
	s.logger.Info("%s", string(body))
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
