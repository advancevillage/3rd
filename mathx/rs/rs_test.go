package rs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_rs_ggx(t *testing.T) {
	var data = map[string]struct {
		m     uint32
		nEcCw uint32
		alpha []uint32
	}{
		"m=8 nEcCw=7": {
			m:     8,
			nEcCw: 7,
			alpha: []uint32{21, 102, 238, 149, 146, 229, 87, 0},
		},
		"m=8 nEcCw=10": {
			m:     8,
			nEcCw: 10,
			alpha: []uint32{45, 32, 94, 64, 70, 118, 61, 46, 67, 251, 0},
		},
		"m=8 nEcCw=68": {
			m:     8,
			nEcCw: 68,
			alpha: []uint32{238, 163, 8, 5, 3, 127, 184, 101, 27, 235, 238, 43, 198, 175, 215, 82, 32, 54, 2, 118, 225, 166, 241, 137, 125, 41, 177, 52, 231, 95, 97, 199, 52, 227, 89, 160, 173, 253, 84, 15, 84, 93, 151, 203, 220, 165, 202, 60, 52, 133, 205, 190, 101, 84, 150, 43, 254, 32, 160, 90, 70, 77, 93, 224, 33, 223, 159, 247, 0},
		},
	}

	for n, p := range data {
		f := func(t *testing.T) {
			r, err := newrs(p.m)
			if err != nil {
				t.Fatal(err)
				return
			}
			act := r.ggx(p.nEcCw)

			for i := 0; i < len(act); i++ {
				act[i] = r.g.Ptoa(act[i])
			}

			assert.Equal(t, p.alpha, act)
		}
		t.Run(n, f)
	}

}

func Test_rs_encode(t *testing.T) {
	var data = map[string]struct {
		m     uint32
		nEcCw uint32
		cw    []uint32
		exp   []uint32
	}{
		"m=4 nEcCw=4": {
			// gx =  [12 1 3 15 1]
			// mx =  [0 0 0 0 11 10 9 8 7 6 5 4 3 2 1]
			// i =  10  mx =  [0 0 0 0 11 10 9 8 7 6 5 4 3 2 1]
			// i =  9   mx =  [0 0 0 0 11 10 9 8 7 6 9 5 0 13 0]
			// i =  8   mx =  [0 0 0 0 11 10 9 8 7 5 4 1 7 0 0]
			// i =  7   mx =  [0 0 0 0 11 10 9 8 5 2 13 10 0 0 0]
			// i =  6   mx =  [0 0 0 0 11 10 9 9 15 15 1 0 0 0 0]
			// i =  5   mx =  [0 0 0 0 11 10 5 8 12 0 0 0 0 0 0]
			// i =  4   mx =  [0 0 0 0 11 10 5 8 12 0 0 0 0 0 0]
			// i =  3   mx =  [0 0 0 0 4 6 2 0 0 0 0 0 0 0 0]
			// i =  2   mx =  [0 0 0 0 4 6 2 0 0 0 0 0 0 0 0]
			// i =  1   mx =  [0 0 11 2 2 11 0 0 0 0 0 0 0 0 0]
			// i =  0   mx =  [0 13 0 12 1 0 0 0 0 0 0 0 0 0 0]
			m:     4,
			nEcCw: 4,
			cw:    []uint32{11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1},
			exp:   []uint32{12, 12, 3, 3, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1},
		},
		"m=8 nEcCw=10": {
			m:     8,
			nEcCw: 10,
			cw:    []uint32{17, 236, 17, 236, 17, 236, 64, 67, 77, 220, 114, 209, 120, 11, 91, 32},
			exp:   []uint32{23, 93, 226, 231, 215, 235, 119, 39, 35, 196, 17, 236, 17, 236, 17, 236, 64, 67, 77, 220, 114, 209, 120, 11, 91, 32},
		},
	}

	for n, p := range data {
		f := func(t *testing.T) {
			r, err := newrs(p.m)
			if err != nil {
				t.Fatal(err)
				return
			}
			act := r.encode(p.cw, p.nEcCw)

			assert.Equal(t, p.exp, act)
		}
		t.Run(n, f)
	}
}
