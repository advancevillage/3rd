package sts

import "net/http"

type StsOption func(*stsOption)

func WithStsSecret(ak, sk, region string) StsOption {
	return func(o *stsOption) {
		o.ak = ak
		o.sk = sk
		o.region = region
	}
}

func WithStsEndpoint(endpoint string) StsOption {
	return func(o *stsOption) {
		o.endpoint = endpoint
	}
}

func WithStsTimeout(timeout int) StsOption {
	return func(o *stsOption) {
		o.timeout = timeout
	}
}

func WithStsDuration(durationSec uint64) StsOption {
	return func(o *stsOption) {
		o.defaultDurationSec = durationSec
	}
}

type stsOption struct {
	ak                 string
	sk                 string
	region             string
	endpoint           string
	signMethod         string
	requestMethod      string
	timeout            int
	defaultDurationSec uint64
}

var defaultStsOption = stsOption{
	endpoint:           "sts.tencentcloudapi.com",
	signMethod:         "TC3-HMAC-SHA256",
	requestMethod:      http.MethodPost,
	timeout:            30,
	defaultDurationSec: 1800,
}
