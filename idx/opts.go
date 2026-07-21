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

type SearchOption = x.Options[reqOption]

func WithLimit(limit int) SearchOption {
	return x.NewFuncOptions(func(o *reqOption) {
		o.limit = limit
	})
}

func WithMatchThreshold(threshold int) SearchOption {
	return x.NewFuncOptions(func(o *reqOption) {
		o.matchThreshold = threshold
	})
}

func WithFilter(filter Filter) SearchOption {
	return x.NewFuncOptions(func(o *reqOption) {
		o.filter = filter
	})
}

func WithMode(mode SearchMode) SearchOption {
	return x.NewFuncOptions(func(o *reqOption) {
		o.mode = mode
	})
}

type reqOption struct {
	mode           SearchMode
	limit          int
	filter         Filter
	matchThreshold int
}

var defaultSearchOption = reqOption{
	mode:           ModeText,
	limit:          2,
	matchThreshold: 75,
}
