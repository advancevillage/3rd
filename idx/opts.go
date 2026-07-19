package idx

import (
	"fmt"
	"net/http"
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

func WithSessionToken(token string) Option {
	return x.NewFuncOptions(func(o *option) {
		o.token = token
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

func WithEndpoint(endpoint string) Option {
	return x.NewFuncOptions(func(o *option) {
		o.endpoint = endpoint
	})
}

func WithTimeout(timeout time.Duration) Option {
	return x.NewFuncOptions(func(o *option) {
		o.timeout = timeout
	})
}

func WithHTTPClient(c *http.Client) Option {
	return x.NewFuncOptions(func(o *option) {
		o.httpClient = c
	})
}

type option struct {
	ak         string
	sk         string
	token      string
	appId      string
	region     string
	endpoint   string
	timeout    time.Duration
	httpClient *http.Client
}

var defaultOption = option{
	timeout: 30 * time.Second,
}

func (o option) validate() error {
	if o.ak == "" || o.sk == "" {
		return fmt.Errorf("idx: invalid ak or sk")
	}
	if o.appId == "" {
		return fmt.Errorf("idx: invalid appid")
	}
	if !validRegion(o.region) {
		return fmt.Errorf("idx: unsupported region: %s", o.region)
	}
	return nil
}

func validRegion(region string) bool {
	switch region {
	case "ap-beijing", "ap-shanghai", "ap-chengdu":
		return true
	default:
		return false
	}
}
