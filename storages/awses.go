//author: richard
package storages

import (
	"github.com/advancevillage/3rd/logs"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elasticsearchservice"
)

func NewAwsES(ak string, sk string, region string, endpoint string, logger logs.Logs) (*AwsES, error) {
	awsEs := AwsES{}
	awsEs.logger = logger
	s, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(ak, sk, ""),
		Endpoint: aws.String(endpoint),
		Region: aws.String(region),
		DisableSSL:  aws.Bool(true),
	})
	if err != nil {
		awsEs.logger.Error(err.Error())
		return nil, err
	}
	awsEs.es = elasticsearchservice.New(s)
	return &awsEs, nil
}


