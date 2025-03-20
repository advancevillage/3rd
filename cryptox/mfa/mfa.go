package mfa

import (
	"context"
	"fmt"
	"strings"

	"github.com/advancevillage/3rd/logx"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

type MFA interface {
	Valid(ctx context.Context, otp string) bool
}

var _ MFA = (*mfa)(nil)

type mfa struct {
	key    *otp.Key
	logger logx.ILogger
}

func NewMFA(ctx context.Context, logger logx.ILogger, opt ...MFAOption) (MFA, error) {
	return newMFA(ctx, logger, opt...)
}

func newMFA(ctx context.Context, logger logx.ILogger, opt ...MFAOption) (*mfa, error) {
	// 1. 解析配置
	opts := defaultMFAOptions
	for _, o := range opt {
		o.apply(&opts)
	}
	// 2. 解析域名
	parts := strings.Split(opts.email, "@")
	if len(parts) != 2 {
		logger.Errorw(ctx, "failed to auth email", "email", opts.email)
		return nil, fmt.Errorf("failed to auth email %s", opts.email)
	}
	// 3. 生成密钥
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      parts[1],
		AccountName: opts.email,
		Secret:      []byte(opts.sk),
		Period:      opts.period,
	})
	if err != nil {
		logger.Errorw(ctx, "failed to generate key", "err", err)
		return nil, err
	}
	logger.Infow(ctx, "success to generate mfa", "qr", key.String())
	return &mfa{key: key, logger: logger}, nil
}

func (m *mfa) Valid(ctx context.Context, otp string) bool {
	return totp.Validate(otp, m.key.Secret())
}
