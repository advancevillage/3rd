package x_test

import (
	"testing"

	"github.com/advancevillage/3rd/mathx"
	"github.com/advancevillage/3rd/x"
	"github.com/stretchr/testify/assert"
)

func Test_builder(t *testing.T) {
	var data = map[string]struct {
		key   string
		value interface{}
	}{
		"case1": {
			key:   mathx.RandStr(10),
			value: mathx.RandStrNum(100),
		},
	}
	for n, v := range data {
		f := func(t *testing.T) {
			exp := map[string]interface{}{
				v.key: v.value,
			}
			t.Log(v.key, v.value)
			b := x.NewBuilder(x.WithKV(v.key, v.value))
			assert.Equal(t, exp, b.Build())
		}
		t.Run(n, f)
	}

}
