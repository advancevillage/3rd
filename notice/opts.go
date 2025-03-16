package notice

import "net/http"

type NoticeOption interface {
	apply(*noticeOption)
}

func WithSmsSecret(ak, sk, region string) NoticeOption {
	return newFuncNoticeOption(func(o *noticeOption) {
		o.ak = ak
		o.sk = sk
		o.region = region
	})
}

func WithSmsApp(appId, signName, tmplId string) NoticeOption {
	return newFuncNoticeOption(func(o *noticeOption) {
		o.smsAppId = appId
		o.smsTmplId = tmplId
		o.smsSignName = signName
	})
}

type noticeOption struct {
	ak            string // 腾讯云短信密钥
	sk            string //腾讯云短信密钥
	region        string // 腾讯云短信密钥
	timeout       int    // 腾讯云短信超时时间
	endpoint      string // 腾讯云短信请求地址
	signMethod    string // 腾讯云短信签名方法
	requestMethod string // 腾讯云短信请求方法

	smsAppId    string // 腾讯云短信应用ID
	smsSignName string // 腾讯云短信签名
	smsTmplId   string // 腾讯云短信模板ID
}

var defaultNoticeOptions = noticeOption{
	timeout:       30,
	endpoint:      "sms.tencentcloudapi.com",
	signMethod:    "TC3-HMAC-SHA256",
	requestMethod: http.MethodPost,
}

type funcNoticeOption struct {
	f func(*noticeOption)
}

func (fdo *funcNoticeOption) apply(do *noticeOption) {
	fdo.f(do)
}

func newFuncNoticeOption(f func(*noticeOption)) *funcNoticeOption {
	return &funcNoticeOption{
		f: f,
	}
}
