package mathx_test

import (
	"testing"

	"github.com/advancevillage/3rd/mathx"
	"github.com/stretchr/testify/assert"
)

func Test_SmoothShiftedSigmoid(t *testing.T) {
	var data = map[string]struct {
		alpha float64
		x0    float64
		min   float64
		max   float64
	}{
		"case1": {
			alpha: 0.08,
			x0:    60,
			min:   0,
			max:   120,
		},
	}

	for n, p := range data {
		f := func(t *testing.T) {
			for i := p.min; i < p.max; i++ {
				v := mathx.SmoothShiftedSigmoid(i, p.alpha, p.x0)
				t.Logf("x=%0.0f y=%d%%", i, int(v*100))
				assert.Equal(t, true, v < 1.0 && v > 0.0)
			}
		}
		t.Run(n, f)
	}
}
