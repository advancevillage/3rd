package mathx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var randTest = map[string]struct {
	n  int
	nn int
}{
	"case-10k": {
		n:  10 * 1000,
		nn: 8,
	},
	"case-20k": {
		n:  20 * 1000,
		nn: 64,
	},
	"case-100k": {
		n:  100 * 1000,
		nn: 128,
	},
	"case-1000k": {
		n:  1000 * 1000,
		nn: 64,
	},
}

func Test_rand(t *testing.T) {
	for n, p := range randTest {
		f := func(t *testing.T) {
			var r = make(map[string]struct{}, p.n)
			for i := 0; i < p.n; i++ {
				r[RandNum(p.nn)] = struct{}{}
			}
			assert.Equal(t, p.n, len(r))

			var rr = make(map[string]struct{}, p.n)
			for i := 0; i < p.n; i++ {
				rr[RandStr(p.nn)] = struct{}{}
			}
			assert.Equal(t, p.n, len(rr))

			var rrr = make(map[string]struct{}, p.n)
			for i := 0; i < p.n; i++ {
				rrr[RandStrNum(p.nn)] = struct{}{}
			}
			assert.Equal(t, p.n, len(rrr))
		}
		t.Run(n, f)
	}

}
