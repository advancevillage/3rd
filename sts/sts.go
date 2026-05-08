package sts

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/advancevillage/3rd/logx"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	sts "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sts/v20180813"
)

type ISts interface {
	Issue(ctx context.Context, name, policy string, durationSec uint64) (string, error)
}

var _ ISts = (*txsts)(nil)

type txsts struct {
	opts   stsOption
	logger logx.ILogger
	client *sts.Client
}

// sts://ak:sk@region?endpoint=sts.tencentcloudapi.com&duration=1800
func NewStsClient(ctx context.Context, logger logx.ILogger, dsn string) (ISts, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return nil, err
	}
	if u.Scheme != "sts" {
		return nil, fmt.Errorf("invalid scheme: %s", u.Scheme)
	}
	ak := u.User.Username()
	sk, ok := u.User.Password()
	if !ok || len(ak) <= 0 || len(sk) <= 0 {
		return nil, fmt.Errorf("invalid ak or sk")
	}
	region := u.Host
	if len(region) <= 0 {
		return nil, fmt.Errorf("invalid region")
	}
	opts := []StsOption{WithStsSecret(ak, sk, region)}
	if v := u.Query().Get("endpoint"); len(v) > 0 {
		opts = append(opts, WithStsEndpoint(v))
	}
	if v := u.Query().Get("duration"); len(v) > 0 {
		d, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid duration: %s", v)
		}
		opts = append(opts, WithStsDuration(d))
	}
	return newTxSts(ctx, logger, opts...)
}

func NewTxSts(ctx context.Context, logger logx.ILogger, opt ...StsOption) (ISts, error) {
	return newTxSts(ctx, logger, opt...)
}

func newTxSts(ctx context.Context, logger logx.ILogger, opt ...StsOption) (*txsts, error) {
	// 1. 配置
	opts := defaultStsOption
	for _, o := range opt {
		o.Apply(&opts)
	}

	// 2. 创建凭证
	credential := common.NewCredential(opts.ak, opts.sk)

	// 3. 客户端配置对象
	cpf := profile.NewClientProfile()
	cpf.SignMethod = opts.signMethod
	cpf.HttpProfile.Endpoint = opts.endpoint
	cpf.HttpProfile.ReqMethod = http.MethodPost
	cpf.HttpProfile.ReqTimeout = opts.timeout

	// 4. 创建客户端
	c, err := sts.NewClient(credential, opts.region, cpf)
	if err != nil {
		logger.Errorw(ctx, "failed to create tx sts client", "err", err)
		return nil, err
	}

	// 5. 返回
	return &txsts{
		opts:   opts,
		logger: logger,
		client: c,
	}, nil
}

// 官方文档
// https://cloud.tencent.com/document/product/598/13896
func (c *txsts) Issue(ctx context.Context, name, policy string, durationSec uint64) (string, error) {
	// 1. 校验参数
	if len(name) <= 0 {
		return "", fmt.Errorf("invalid name")
	}
	if len(policy) <= 0 {
		return "", fmt.Errorf("invalid policy")
	}
	if durationSec <= 0 {
		durationSec = c.opts.defaultDurationSec
	}

	// 2. 构建请求
	req := sts.NewGetFederationTokenRequest()
	req.Name = common.StringPtr(name)
	req.Policy = common.StringPtr(policy)
	req.DurationSeconds = common.Uint64Ptr(durationSec)

	// 3. 调用
	reply, err := c.client.GetFederationTokenWithContext(ctx, req)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		c.logger.Errorw(ctx, "failed to issue sts by sdk", "err", err, "name", name)
		return "", err
	}
	if err != nil {
		c.logger.Errorw(ctx, "failed to issue sts", "err", err, "name", name)
		return "", err
	}

	// 4. 校验返回
	if reply.Response == nil || reply.Response.Credentials == nil {
		c.logger.Errorw(ctx, "failed to issue sts: empty credentials", "name", name)
		return "", fmt.Errorf("empty credentials")
	}
	cred := reply.Response.Credentials
	if cred.TmpSecretId == nil || cred.TmpSecretKey == nil || cred.Token == nil {
		c.logger.Errorw(ctx, "failed to issue sts: incomplete credentials", "name", name)
		return "", fmt.Errorf("incomplete credentials")
	}
	var exp int64
	if reply.Response.ExpiredTime != nil {
		exp = int64(*reply.Response.ExpiredTime)
	}

	// 5. 编码 DSN
	dsn := encodeCredential(*cred.TmpSecretId, *cred.TmpSecretKey, *cred.Token, c.opts.region, exp)

	// 6. 日志打点（不含密钥/Token）
	var requestId string
	if reply.Response.RequestId != nil {
		requestId = *reply.Response.RequestId
	}
	c.logger.Infow(ctx, "issue sts success", "name", name, "request_id", requestId, "expired_time", exp)

	return dsn, nil
}

func encodeCredential(tmpId, tmpKey, token, region string, exp int64) string {
	u := &url.URL{
		Scheme: "sts",
		User:   url.UserPassword(tmpId, tmpKey),
		Host:   region,
	}
	q := url.Values{}
	q.Set("token", token)
	q.Set("exp", strconv.FormatInt(exp, 10))
	u.RawQuery = q.Encode()
	return u.String()
}
