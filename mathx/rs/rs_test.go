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
		"m=8 nEcCw = 7": {
			m:     8,
			nEcCw: 7,
			alpha: []uint32{21, 102, 238, 149, 146, 229, 87, 0},
		},
		"m=8 nEcCw = 10": {
			m:     8,
			nEcCw: 10,
			alpha: []uint32{45, 32, 94, 64, 70, 118, 61, 46, 67, 251, 0},
		},
		"m=8 nEcCw = 68": {
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
