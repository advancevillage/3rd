package idx

import (
	"time"

	"github.com/advancevillage/3rd/x"
)

type Option = x.Options[option]

func WithSecret(ak, sk string) Option {
	return x.NewFuncOptions(func(o *option) {
		o.ak = ak
		o.sk = sk
	})
}

func WithAppId(appId string) Option {
	return x.NewFuncOptions(func(o *option) {
		o.appId = appId
	})
}

func WithRegion(region string) Option {
	return x.NewFuncOptions(func(o *option) {
		o.region = region
	})
}

func WithTimeout(timeout time.Duration) Option {
	return x.NewFuncOptions(func(o *option) {
		o.timeout = timeout
	})
}

type option struct {
	ak      string
	sk      string
	appId   string
	region  string
	timeout time.Duration
}

var defaultOption = option{
	region:  "ap-shanghai",
	timeout: 30 * time.Second,
}
