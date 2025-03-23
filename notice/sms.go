package notice

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/advancevillage/3rd/logx"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	sms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
)

type SMS interface {
	Send(ctx context.Context, phone string, args ...string) error
}

var _ SMS = (*txsms)(nil)

type txsms struct {
	opts   noticeOption
	logger logx.ILogger
	client *sms.Client
}

func NewTxSms(ctx context.Context, logger logx.ILogger, opt ...NoticeOption) (SMS, error) {
	return newTxSms(ctx, logger, opt...)
}

func newTxSms(ctx context.Context, logger logx.ILogger, opt ...NoticeOption) (*txsms, error) {
	// 1. 配置
	opts := defaultNoticeOptions
	for _, o := range opt {
		o.apply(&opts)
	}

	// 2. 创建凭证
	credential := common.NewCredential(opts.ak, opts.sk)

	// 3. 客户端配置对象
	cpf := profile.NewClientProfile()
	cpf.SignMethod = opts.signMethod
	cpf.HttpProfile.Endpoint = opts.endpoint
	cpf.HttpProfile.ReqMethod = opts.requestMethod
	cpf.HttpProfile.ReqTimeout = opts.timeout

	// 4. 创建客户端
	c, err := sms.NewClient(credential, opts.region, cpf)
	if err != nil {
		logger.Errorw(ctx, "failed to create tx sms client", "err", err)
		return nil, err
	}

	// 5. 返回
	return &txsms{
		opts:   opts,
		logger: logger,
		client: c,
	}, nil
}

// 官方文档
// https://cloud.tencent.com/document/product/382/55981
func (c *txsms) Send(ctx context.Context, phone string, args ...string) error {
	// 1. 构建请请求
	if !strings.HasPrefix(phone, "+86") {
		phone = "+86" + phone
	}

	// 2. 请求
	req := sms.NewSendSmsRequest()

	// 3. 基本配置
	req.SmsSdkAppId = common.StringPtr(c.opts.smsAppId)
	req.SignName = common.StringPtr(c.opts.smsSignName)
	req.TemplateId = common.StringPtr(c.opts.smsTmplId)

	// 4. 云平台模版参数
	req.TemplateParamSet = common.StringPtrs(args)
	req.PhoneNumberSet = common.StringPtrs([]string{phone})

	// 5. 扩展参数
	req.SessionContext = common.StringPtr("")

	// 6. 发送
	reply, err := c.client.SendSms(req)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		return nil
	}
	if err != nil {
		c.logger.Errorw(ctx, "failed to send sms", "err", err, "phone", phone)
		return err
	}
	// 7. 日志打点
	b, err := json.Marshal(reply.Response)
	if err != nil {
		c.logger.Errorw(ctx, "failed to marshal tx sms response", "err", err, "phone", phone)
		return err
	}
	c.logger.Infow(ctx, "send sms success", "phone", phone, "reply", string(b))
	return nil
}
