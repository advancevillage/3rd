package mfa_test

import (
	"context"
	"testing"

	"github.com/advancevillage/3rd/cryptox/mfa"
	"github.com/advancevillage/3rd/logx"
	"github.com/advancevillage/3rd/mathx"
	"github.com/stretchr/testify/assert"
)

func Test_mfa(t *testing.T) {
	ctx := context.WithValue(context.TODO(), logx.TraceId, mathx.UUID())
	logger, err := logx.NewLogger("debug")
	assert.Nil(t, err)
	c, err := mfa.NewMFA(ctx, logger, mfa.WithSecret("richard.sun"))
	assert.Nil(t, err)
	assert.Equal(t, false, c.Valid(ctx, mathx.RandNum(6)))
}
