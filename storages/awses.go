//author: richard
package storages

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/advancevillage/3rd/logs"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/signer/v4"
	"github.com/aws/aws-sdk-go/service/elasticsearchservice"
	"io/ioutil"
	"net/http"
	"time"
)

func NewAwsES(ak string, sk string, region string, domain string, logger logs.Logs) *AwsES {
	es := AwsES{}
	es.logger = logger
	es.ak = ak
	es.sk = sk
	es.domain = domain
	es.region = region
	es.service = "es"
	return &es
}

func (s *AwsES) DeleteStorage(index string, key string) error {
	endpoint := fmt.Sprintf("%s/%s/_doc/%s", s.domain, index, key)
	//sign v4
	credential := credentials.NewStaticCredentials(s.ak, s.sk, "")
	signer := v4.NewSigner(credential)
	buf := bytes.NewReader(nil)
	//调用ES API
	client := &http.Client{}
	request, err := http.NewRequest(http.MethodDelete, endpoint, buf)
	if err != nil {
		s.logger.Error(err.Error())
		return err
	}
	request.Header.Add("Content-Type", "application/json")
	_, err = signer.Sign(request, buf, s.service, s.region, time.Now())
	if err != nil {
		s.logger.Error(err.Error())
		return err
	}
	response, err := client.Do(request)
	if err != nil {
		s.logger.Error(err.Error())
		return err
	}
	if response.StatusCode != http.StatusOK {
		return errors.New(elasticsearchservice.ErrCodeBaseException)
	}
	return nil
}

func (s *AwsES) CreateStorageV2(index string, key string, body []byte) error {
	endpoint := fmt.Sprintf("%s/%s/_doc/%s", s.domain, index, key)
	//sign v4
	credential := credentials.NewStaticCredentials(s.ak, s.sk, "")
	signer := v4.NewSigner(credential)
	buf := bytes.NewReader(body)
	//调用ES API
	client := &http.Client{}
	request, err := http.NewRequest(http.MethodPost, endpoint, buf)
	if err != nil {
		s.logger.Error(err.Error())
		return err
	}
	request.Header.Add("Content-Type", "application/json")
	_, err = signer.Sign(request, buf, s.service, s.region, time.Now())
	if err != nil {
		s.logger.Error(err.Error())
		return err
	}
	_, err = client.Do(request)
	if err != nil {
		s.logger.Error(err.Error())
		return err
	}
	return nil
}

func (s *AwsES) UpdateStorageV2(index string, key string, body []byte) error {
	endpoint := fmt.Sprintf("%s/%s/_doc/%s", s.domain, index, key)
	//sign v4
	credential := credentials.NewStaticCredentials(s.ak, s.sk, "")
	signer := v4.NewSigner(credential)
	buf := bytes.NewReader(body)
	//调用ES API
	client := &http.Client{}
	request, err := http.NewRequest(http.MethodPut, endpoint, buf)
	if err != nil {
		s.logger.Error(err.Error())
		return err
	}
	request.Header.Add("Content-Type", "application/json")
	_, err = signer.Sign(request, buf, s.service, s.region, time.Now())
	if err != nil {
		s.logger.Error(err.Error())
		return err
	}
	response, err := client.Do(request)
	if err != nil {
		s.logger.Error(err.Error())
		return err
	}
	if response.StatusCode != http.StatusOK {
		return errors.New(elasticsearchservice.ErrCodeBaseException)
	}
	return nil
}

func (s *AwsES) DeleteStorageV2(index string, key ...string) error {
	for i := range key {
		err := s.DeleteStorage(index, key[i])
		if err != nil {
			s.logger.Error(err.Error())
		} else {
			continue
		}
	}
	return nil
}

func (s *AwsES) QueryStorageV2(index string, key  string) ([]byte, error) {
	endpoint := fmt.Sprintf("%s/%s/_doc/%s", s.domain, index, key)
	//sign v4
	credential := credentials.NewStaticCredentials(s.ak, s.sk, "")
	signer := v4.NewSigner(credential)
	buf := bytes.NewReader(nil)
	//调用ES API
	client := &http.Client{}
	request, err := http.NewRequest(http.MethodGet, endpoint, buf)
	if err != nil {
		s.logger.Error(err.Error())
		return nil, err
	}
	request.Header.Add("Content-Type", "application/json")
	_, err = signer.Sign(request, buf, s.service, s.region, time.Now())
	if err != nil {
		s.logger.Error(err.Error())
		return nil, err
	}
	response, err := client.Do(request)
	if err != nil {
		s.logger.Error(err.Error())
		return nil, err
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		s.logger.Error(err.Error())
		return nil, err
	}
	return body, nil
}













