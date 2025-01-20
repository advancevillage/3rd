package mathx_test

import (
	"testing"

	"github.com/advancevillage/3rd/mathx"
	"github.com/stretchr/testify/assert"
)

func Test_gid(t *testing.T) {
	var data = map[string]struct {
		count int
	}{
		"case-10": {
			count: 10,
		},
		"case-10k": {
			count: 10000,
		},
		"case-1m": {
			count: 1000000,
		},
	}

	for n, v := range data {
		f := func(t *testing.T) {
			m := make(map[int64]struct{})
			for i := 0; i < v.count; i++ {
				gid := mathx.GId()
				m[gid] = struct{}{}
			}
			assert.Equal(t, v.count, len(m))
		}
		t.Run(n, f)
	}
}
