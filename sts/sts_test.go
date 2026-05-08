package sts_test

import (
	"context"
	"os"
	"testing"

	"github.com/advancevillage/3rd/logx"
	"github.com/advancevillage/3rd/mathx"
	"github.com/advancevillage/3rd/sts"
	"github.com/stretchr/testify/assert"
)

func newTestLogger(t *testing.T) (logx.ILogger, context.Context) {
	t.Helper()
	logger, err := logx.NewLogger("debug")
	assert.Nil(t, err)
	ctx := context.WithValue(context.TODO(), logx.TraceId, mathx.UUID())
	return logger, ctx
}

func Test_tx_sts(t *testing.T) {
	logger, ctx := newTestLogger(t)
	dsn := os.Getenv("STS_DSN")

	data := map[string]struct {
		name        string
		policy      string
		durationSec uint64
	}{
		"case1": {
			name:        "3rd-test",
			policy:      `{"version":"2.0","statement":[{"effect":"allow","action":["asr:*"],"resource":["*"]}]}`,
			durationSec: 60,
		},
	}

	for n, v := range data {
		f := func(t *testing.T) {
			c, err := sts.NewStsClient(ctx, logger, dsn)
			assert.Nil(t, err)
			iss, err := c.Issue(ctx, v.name, v.policy, v.durationSec)
			assert.Nil(t, err)
			assert.NotEmpty(t, iss)
			t.Logf("issue: %s", iss)
		}
		t.Run(n, f)
	}
}
