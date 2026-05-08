package sts

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"testing"

	"github.com/advancevillage/3rd/logx"
	"github.com/advancevillage/3rd/mathx"
	"github.com/stretchr/testify/assert"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
)

func newTestLogger(t *testing.T) (logx.ILogger, context.Context) {
	t.Helper()
	logger, err := logx.NewLogger("debug")
	assert.Nil(t, err)
	ctx := context.WithValue(context.TODO(), logx.TraceId, mathx.UUID())
	return logger, ctx
}

func Test_NewStsClient_dsn(t *testing.T) {
	logger, ctx := newTestLogger(t)

	t.Run("invalid scheme", func(t *testing.T) {
		_, err := NewStsClient(ctx, logger, "http://ak:sk@ap-guangzhou")
		assert.NotNil(t, err)
	})

	t.Run("missing ak/sk", func(t *testing.T) {
		_, err := NewStsClient(ctx, logger, "sts://@ap-guangzhou")
		assert.NotNil(t, err)
	})

	t.Run("missing region", func(t *testing.T) {
		_, err := NewStsClient(ctx, logger, "sts://ak:sk@")
		assert.NotNil(t, err)
	})

	t.Run("invalid duration", func(t *testing.T) {
		_, err := NewStsClient(ctx, logger, "sts://ak:sk@ap-guangzhou?duration=abc")
		assert.NotNil(t, err)
	})

	t.Run("ok with overrides", func(t *testing.T) {
		dsn := "sts://ak:sk@ap-guangzhou?endpoint=sts.internal.tencentcloudapi.com&duration=900"
		c, err := NewStsClient(ctx, logger, dsn)
		assert.Nil(t, err)
		assert.NotNil(t, c)
		impl, ok := c.(*txsts)
		assert.True(t, ok)
		assert.Equal(t, "ak", impl.opts.ak)
		assert.Equal(t, "sk", impl.opts.sk)
		assert.Equal(t, "ap-guangzhou", impl.opts.region)
		assert.Equal(t, "sts.internal.tencentcloudapi.com", impl.opts.endpoint)
		assert.Equal(t, uint64(900), impl.opts.defaultDurationSec)
	})

	t.Run("ok with defaults", func(t *testing.T) {
		c, err := NewStsClient(ctx, logger, "sts://ak:sk@ap-guangzhou")
		assert.Nil(t, err)
		impl := c.(*txsts)
		assert.Equal(t, "sts.tencentcloudapi.com", impl.opts.endpoint)
		assert.Equal(t, uint64(1800), impl.opts.defaultDurationSec)
		assert.Equal(t, "TC3-HMAC-SHA256", impl.opts.signMethod)
		assert.Equal(t, 30, impl.opts.timeout)
	})
}

func Test_encodeCredential(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		dsn := encodeCredential("tid", "tkey", "tok", "ap-guangzhou", 1700000000)
		u, err := url.Parse(dsn)
		assert.Nil(t, err)
		assert.Equal(t, "sts", u.Scheme)
		assert.Equal(t, "ap-guangzhou", u.Host)
		assert.Equal(t, "tid", u.User.Username())
		pwd, ok := u.User.Password()
		assert.True(t, ok)
		assert.Equal(t, "tkey", pwd)
		assert.Equal(t, "tok", u.Query().Get("token"))
		exp, err := strconv.ParseInt(u.Query().Get("exp"), 10, 64)
		assert.Nil(t, err)
		assert.Equal(t, int64(1700000000), exp)
	})

	t.Run("token with special chars", func(t *testing.T) {
		token := "abc+def/ghi=jkl&mno"
		dsn := encodeCredential("tid", "tkey", token, "ap-guangzhou", 0)
		u, err := url.Parse(dsn)
		assert.Nil(t, err)
		assert.Equal(t, token, u.Query().Get("token"))
	})
}

func Test_tx_sts_Issue_invalid_args(t *testing.T) {
	logger, ctx := newTestLogger(t)
	c, err := NewStsClient(ctx, logger, "sts://ak:sk@ap-guangzhou")
	assert.Nil(t, err)

	_, err = c.Issue(ctx, "", "{}", 0)
	assert.NotNil(t, err)

	_, err = c.Issue(ctx, "n1", "", 0)
	assert.NotNil(t, err)
}

func Test_tx_sts_Issue_sdk_error(t *testing.T) {
	logger, ctx := newTestLogger(t)
	c, err := NewStsClient(ctx, logger, "sts://ak:sk@ap-guangzhou")
	assert.Nil(t, err)

	_, err = c.Issue(ctx, "peeko-test", `{"version":"2.0","statement":[{"effect":"allow","action":["*"],"resource":["*"]}]}`, 900)
	assert.NotNil(t, err)
	_, ok := err.(*errors.TencentCloudSDKError)
	assert.True(t, ok, "expect tencent sdk error, got %T", err)
}

func Test_tx_sts_Issue_real(t *testing.T) {
	ak := os.Getenv("STS_AK")
	sk := os.Getenv("STS_SK")
	rgn := os.Getenv("STS_RGN")
	if len(ak) <= 0 || len(sk) <= 0 || len(rgn) <= 0 {
		t.Skip("STS_AK/STS_SK/STS_RGN not set")
	}
	logger, ctx := newTestLogger(t)
	dsn := fmt.Sprintf("sts://%s:%s@%s", ak, sk, rgn)
	c, err := NewStsClient(ctx, logger, dsn)
	assert.Nil(t, err)
	policy := `{"version":"2.0","statement":[{"effect":"allow","action":["name/sts:GetFederationToken"],"resource":["*"]}]}`
	out, err := c.Issue(ctx, "peeko-test", policy, 900)
	assert.Nil(t, err)
	t.Log(out)
	u, err := url.Parse(out)
	assert.Nil(t, err)
	assert.Equal(t, "sts", u.Scheme)
}
