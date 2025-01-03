package dbx

import (
	"context"
	"net/http"
	"time"

	"github.com/tencentyun/cos-go-sdk-v5"
)

type S3 interface {
}

var _ S3 = (*TxCos)(nil)

type TxCos struct {
	c  *cos.Client
	ak string
	sk string
}

func NewCosS3(ctx context.Context, bucket string, region string, ak string, sk string) (S3, error) {
	b, err := cos.NewBucketURL(bucket, region, true)
	if err != nil {
		return nil, err
	}
	c := cos.NewClient(&cos.BaseURL{BucketURL: b}, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  ak,
			SecretKey: sk,
		},
	})
	return &TxCos{c: c, ak: ak, sk: sk}, nil
}

func (t *TxCos) Download(ctx context.Context, name string) (string, error) {
	return t.getPresignedUrl(ctx, http.MethodGet, name)
}

func (t *TxCos) Exist(ctx context.Context, name string) (bool, error) {
	_, err := t.c.Object.Head(ctx, name, nil)
	switch {
	case err == nil:
		return true, nil

	case cos.IsNotFoundError(err):
		return false, nil

	default:
		return false, err
	}
}

func (t *TxCos) getPresignedUrl(ctx context.Context, httpMethod string, name string) (string, error) {
	url, err := t.c.Object.GetPresignedURL(ctx, httpMethod, name, t.ak, t.sk, time.Hour, nil)
	if err != nil {
		return "", err
	}
	return url.String(), nil
}
