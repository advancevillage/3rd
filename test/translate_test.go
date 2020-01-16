//author: richard
package test

import (
	"github.com/advancevillage/3rd/translate"
	"testing"
)

func TestBaiDuTranslate(t *testing.T) {
	baidu := translate.NewBaiDuTranslate("2015063000000001", "12345678")
	str, err := baidu.Translate("richard", "en", "zh")
	if err != nil {
		t.Error(err.Error())
	} else {
		t.Log(str)
	}

}
