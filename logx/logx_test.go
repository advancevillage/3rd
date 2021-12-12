package logx

import (
	"context"
	"testing"

	"github.com/advancevillage/3rd/mathx"
)

var zlogTest = map[string]struct {
	level   string
	traceId string
	msg     string
	key     interface{}
	value   interface{}
}{
	"case-string": {
		level:   "info",
		traceId: mathx.UUID(),
		msg:     "string",
		key:     "key-string",
		value:   "value-string",
	},
	"case-map": {
		level:   "info",
		traceId: mathx.UUID(),
		msg:     "map",
		key:     "key-map",
		value: map[string]interface{}{
			"k1": "v1",
			"k2": 100,
			"k3": []int{1, 2},
			"k4": map[string]interface{}{
				"k41": "v41",
				"k42": 100,
				"k43": []int{1, 2},
			},
		},
	},
}

func Test_zlog(t *testing.T) {
	for n, p := range zlogTest {
		f := func(t *testing.T) {
			var z, err = NewLogger(p.level)
			if err != nil {
				t.Fatal(err)
				return
			}
			var ctx = context.WithValue(context.TODO(), traceId, p.traceId)
			z.Infow(ctx, p.msg, p.key, p.value)
		}
		t.Run(n, f)
	}
}
