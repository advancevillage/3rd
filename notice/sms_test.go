package notice_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/advancevillage/3rd/logx"
	"github.com/advancevillage/3rd/mathx"
	"github.com/advancevillage/3rd/notice"
	"github.com/stretchr/testify/assert"
)

func Test_tx_sms(t *testing.T) {
	logger, err := logx.NewLogger("debug")
	assert.Nil(t, err)
	ctx := context.WithValue(context.TODO(), logx.TraceId, mathx.UUID())
	var (
		ak    = os.Getenv("SMS_AK")
		sk    = os.Getenv("SMS_SK")
		rgn   = os.Getenv("SMS_RGN")
		sign  = os.Getenv("SMS_SIGN")
		tmpId = os.Getenv("SMS_TMP_ID")
		appId = os.Getenv("SMS_APP_ID")
		phone = os.Getenv("SMS_PHONE")
	)
	dsn := fmt.Sprintf("sms://%s:%s@%s?sign=%s&app=%s&tmpl=%s", ak, sk, rgn, sign, appId, tmpId)
	t.Log(dsn)
	ntc, err := notice.NewSmsClient(ctx, logger, dsn)
	assert.Nil(t, err)
	err = ntc.Send(ctx, phone, mathx.RandNum(6), "T1")
	assert.Nil(t, err)
}
