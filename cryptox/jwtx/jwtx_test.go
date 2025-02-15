package jwtx_test

import (
	"context"
	"testing"
	"time"

	"github.com/advancevillage/3rd/cryptox/jwtx"
	"github.com/advancevillage/3rd/logx"
	"github.com/advancevillage/3rd/mathx"
	"github.com/advancevillage/3rd/x"
	"github.com/stretchr/testify/assert"
)

func Test_jwtx(t *testing.T) {
	logger, err := logx.NewLogger("debug")
	assert.Nil(t, err)
	ctx := context.WithValue(context.TODO(), logx.TraceId, mathx.UUID())

	var data = map[string]struct {
		sk  string
		opt []x.Option
		exp time.Duration
	}{
		"case1": {
			sk:  "",
			exp: time.Minute,
			opt: []x.Option{
				x.WithKV("uid", "123456"),
				x.WithKV("role", "admin"),
			},
		},
		"case2": {
			sk:  mathx.RandStrNum(16),
			exp: time.Minute,
			opt: []x.Option{
				x.WithKV("uid", "123456"),
				x.WithKV("info", mathx.RandStr(16)),
			},
		},
	}
	for n, v := range data {
		f := func(t *testing.T) {
			j, err := jwtx.NewJwtXClient(ctx, logger, jwtx.WithSecretKey(v.sk), jwtx.WithExpireTime(v.exp), jwtx.WithAppName(mathx.RandStr(5)))
			assert.Nil(t, err)
			token, err := j.Sign(ctx, x.NewBuilder(v.opt...))
			assert.Nil(t, err)
			assert.Equal(t, true, j.Valid(ctx, token))
			act, err := j.Parse(ctx, token)
			assert.Nil(t, err)
			assert.Equal(t, x.NewBuilder(v.opt...).Build(), act.Build())
			t.Log(token, v.sk)
		}
		t.Run(n, f)
	}
}
