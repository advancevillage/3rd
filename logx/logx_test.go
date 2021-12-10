package logx

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

var zlogTest = map[string]struct {
	level   string
	traceId string
}{
	"case1": {
		level:   "info",
		traceId: uuid.New().String(),
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

			z.Infow(ctx, "msg", "key", "value")

		}
		t.Run(n, f)
	}
}
