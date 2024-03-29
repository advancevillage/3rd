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
	"case-50k": {
		n:  50 * 1000,
		nn: 128,
	},
}

func Test_rand(t *testing.T) {
	for n, p := range randTest {
		f := func(t *testing.T) {
			var r = make(map[string]struct{}, p.n)
			for i := 0; i < p.n; i++ {
				r[RandNum(p.nn)] = struct{}{}
			}
			if p.n-len(r) > 10 {
				assert.Equal(t, p.n, len(r))
			}

			var rr = make(map[string]struct{}, p.n)
			for i := 0; i < p.n; i++ {
				rr[RandStr(p.nn)] = struct{}{}
			}
			if p.n-len(rr) > 10 {
				assert.Equal(t, p.n, len(rr))
			}

			var rrr = make(map[string]struct{}, p.n)
			for i := 0; i < p.n; i++ {
				rrr[RandStrNum(p.nn)] = struct{}{}
			}
			if p.n-len(rrr) > 10 {
				assert.Equal(t, p.n, len(rrr))
			}
		}
		t.Run(n, f)
	}

}

var eratosthenesTest = map[string]struct {
	n   int
	exp []int
}{
	"case1": {
		n:   10,
		exp: []int{2, 3, 5, 7},
	},
	"case2": {
		n:   121,
		exp: []int{2, 3, 5, 7, 11, 13, 17, 19, 23, 29, 31, 37, 41, 43, 47, 53, 59, 61, 67, 71, 73, 79, 83, 89, 97, 101, 103, 107, 109, 113},
	},
}

func Test_eratos(t *testing.T) {
	for n, p := range eratosthenesTest {
		f := func(t *testing.T) {
			act := Primes(p.n)
			assert.Equal(t, p.exp, act)
		}
		t.Run(n, f)
	}
}
