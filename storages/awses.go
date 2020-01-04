//author: richard
package storages

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/advancevillage/3rd/logs"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/signer/v4"
	"github.com/aws/aws-sdk-go/private/protocol/rest"
	"github.com/olivere/elastic/v7"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

func NewAwsES(ak string, sk string, region string, domain string, logger logs.Logs) (*AwsES, error) {
	var err error
	es := AwsES{}
	es.logger = logger
	es.credential = credentials.NewStaticCredentials(ak, sk, "")
	sign :=  &signer{}
	sign.logger = logger
	client := &http.Client{}
	sign.v4 = v4.NewSigner(es.credential)
	sign.service = "es"
	sign.region = region
	sign.transport = client.Transport
	if sign.transport == nil {
		sign.transport = http.DefaultTransport
	}
	client.Transport = sign
	es.conn, err = elastic.NewClient(elastic.SetURL(domain), elastic.SetScheme("https"), elastic.SetHttpClient(client),elastic.SetSniff(false))
	if err != nil {
		es.logger.Error(err.Error())
		return nil, err
	}
	info, code, err := es.conn.Ping(domain).Do(context.Background())
	if err != nil {
		es.logger.Error(err.Error())
		return nil, err
	}
	es.logger.Info("info=%v, code=%d", info, code)
	return &es, nil
}

//实现接口
func (s *AwsES) CreateStorage(key string, body []byte) error {
	var object = make(map[string]interface{})
	var err error
	err = json.Unmarshal(body, &object)
	if err != nil {
		s.logger.Error(err.Error())
		return err
	}
	return s.CreateDocument(ESDefaultIndex, key, object)
}

func (s *AwsES) UpdateStorage(key string, body []byte) error {
	var object = make(map[string]interface{})
	var err error
	err = json.Unmarshal(body, &object)
	if err != nil {
		s.logger.Error(err.Error())
		return err
	}
	return s.CreateDocument(ESDefaultIndex, key, object)
}

func (s *AwsES) DeleteStorage(key ...string) error {
	for i := range key {
		err := s.DeleteDocument(ESDefaultIndex, key[i])
		if err != nil {
			s.logger.Error(err.Error())
		} else {
			continue
		}
	}
	return nil
}

func (s *AwsES) QueryStorage(key string) ([]byte, error) {
	return s.QueryDocument(ESDefaultIndex, key)
}

func (s *AwsES) CreateStorageV2(index string, key string, body []byte) error {
	var object = make(map[string]interface{})
	var err error
	err = json.Unmarshal(body, &object)
	if err != nil {
		s.logger.Error(err.Error())
		return err
	}
	return s.CreateDocument(index, key, object)
}

func (s *AwsES) DeleteStorageV2(index string, key ...string) error {
	for i := range key {
		err := s.DeleteDocument(index, key[i])
		if err != nil {
			s.logger.Error(err.Error())
		} else {
			continue
		}
	}
	return nil
}

func (s *AwsES) UpdateStorageV2(index string, key string, body []byte) error {
	var object = make(map[string]interface{})
	var err error
	err = json.Unmarshal(body, &object)
	if err != nil {
		s.logger.Error(err.Error())
		return err
	}
	return s.CreateDocument(index, key, object)
}

func (s *AwsES) QueryStorageV2(index string, key  string) ([]byte, error) {
	return s.QueryDocument(index, key)
}

//TODO
func (s *AwsES) QueryStorageV3(index string, where map[string]interface{}, limit int, offset int) ([][]byte, error) {
	return nil, nil
}

func (s *AwsES) CreateDocument(index string, id string, body interface{}) error {
	_, err := s.conn.Index().Index(index).Id(id).BodyJson(body).Do(context.Background())
	if err != nil {
		s.logger.Error(err.Error())
		return err
	}
	return nil
}

func (s *AwsES) DeleteDocument(index string, id string) error {
	_, err := s.conn.Delete().Index(index).Id(id).Do(context.Background())
	if err != nil {
		s.logger.Error(err.Error())
		return err
	}
	return nil
}

func (s *AwsES) UpdateDocument(index string, id string, fields map[string]interface{}) error {
	_, err := s.conn.Update().Index(index).Id(id).Doc(fields).Do(context.Background())
	if err != nil {
		s.logger.Error(err.Error())
		return err
	}
	return nil
}

func (s *AwsES) QueryDocument(index string, id string) ([]byte, error) {
	ret , err := s.conn.Get().Index(index).Id(id).Do(context.Background())
	if err != nil {
		s.logger.Error(err.Error())
		return nil, err
	}
	buf, err := json.Marshal(ret.Source)
	if err != nil {
		s.logger.Error(err.Error())
		return nil, err
	}
	return buf, nil
}

//@link: https://github.com/sha1sum/aws_signing_client/blob/master/client.go
type signer struct {
	transport http.RoundTripper
	v4        *v4.Signer
	service   string
	region    string
	logger 	  logs.Logs
}

func (s *signer) RoundTrip(request *http.Request) (*http.Response, error) {
	if h, ok := request.Header["Authorization"]; ok && len(h) > 0 && strings.HasPrefix(h[0], "AWS4") {
		return s.transport.RoundTrip(request)
	}
	request.URL.Scheme = "https"
	if strings.Contains(request.URL.RawPath, "%2C") {
		request.URL.RawPath = rest.EscapePath(request.URL.RawPath, false)
	}
	t := time.Now()
	request.Header.Set("Date", t.Format(time.RFC3339))
	var err error
	switch request.Body {
	case nil:
		_, err = s.v4.Sign(request, nil, s.service, s.region, t)
	default:
		d, err := ioutil.ReadAll(request.Body)
		if err != nil {
			return nil, err
		}
		request.Body = ioutil.NopCloser(bytes.NewReader(d))
		_, err = s.v4.Sign(request, bytes.NewReader(d), s.service, s.region, t)
	}
	if err != nil {
		return nil, err
	}
	response, err := s.transport.RoundTrip(request)
	if err != nil {
		return response, err
	}
	body := "<nil>"
	if response.Body != nil {
		defer func () { _ = response.Body.Close() }()
		buf := new(bytes.Buffer)
		_, err = buf.ReadFrom(response.Body)
		if err != nil {
			return nil, err
		}
		body = buf.String()
		response.Body = ioutil.NopCloser(bytes.NewReader(buf.Bytes()))
	}
	s.logger.Info("Successful response from RoundTripper: %+v\n\nBody: %s\n", response, body)
	return response, nil
}









