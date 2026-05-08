package sts

import (
	"github.com/advancevillage/3rd/x"
)

type StsOption = x.Options[stsOption]

func WithStsSecret(ak, sk, region string) StsOption {
	return x.NewFuncOptions(func(o *stsOption) {
		o.ak = ak
		o.sk = sk
		o.region = region
	})
}

func WithStsEndpoint(endpoint string) StsOption {
	return x.NewFuncOptions(func(o *stsOption) {
		o.endpoint = endpoint
	})
}

func WithStsTimeout(timeout int) StsOption {
	return x.NewFuncOptions(func(o *stsOption) {
		o.timeout = timeout
	})
}

func WithStsDuration(durationSec uint64) StsOption {
	return x.NewFuncOptions(func(o *stsOption) {
		o.defaultDurationSec = durationSec
	})
}

type stsOption struct {
	ak                 string
	sk                 string
	region             string
	endpoint           string
	signMethod         string
	timeout            int
	defaultDurationSec uint64
}

var defaultStsOption = stsOption{
	endpoint:           "sts.tencentcloudapi.com",
	signMethod:         "TC3-HMAC-SHA256",
	timeout:            30,
	defaultDurationSec: 120,
}
