package bch

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_bch_mp(t *testing.T) {
	var data = map[string]struct {
		m   uint32
		t   uint32
		exp [][]uint32
	}{
		"m=4, t=3": {
			m: 4,
			t: 3,
			exp: [][]uint32{
				{0},
				{1, 2, 4, 8},
				{1, 2, 4, 8},
				{3, 6, 12, 9},
				{1, 2, 4, 8},
				{5, 10},
				{3, 6, 12, 9},
				{7, 14, 13, 11},
				{1, 2, 4, 8},
				{3, 6, 12, 9},
				{5, 10},
				{7, 14, 13, 11},
				{3, 6, 12, 9},
				{7, 14, 13, 11},
				{7, 14, 13, 11},
				{15},
			},
		},
	}

	for n, p := range data {
		f := func(t *testing.T) {
			c, err := newbch(p.m, p.t)
			if err != nil {
				t.Fatal(err)
				return
			}

			c.gmp()

			assert.Equal(t, p.exp, c.mp)
		}
		t.Run(n, f)
	}

}

func Test_bch_gx(t *testing.T) {
	var data = map[string]struct {
		m   uint32
		t   uint32
		k   uint32
		exp []uint32
	}{
		"m=4, t=3": {
			m:   4,
			t:   3,
			k:   5,
			exp: []uint32{1, 1, 1, 0, 1, 1, 0, 0, 1, 0, 1},
		},
		"m=4, t=2": {
			m:   4,
			t:   2,
			k:   7,
			exp: []uint32{1, 0, 0, 0, 1, 0, 1, 1, 1},
		},
	}

	for n, p := range data {
		f := func(t *testing.T) {
			c, err := newbch(p.m, p.t)
			if err != nil {
				t.Fatal(err)
				return
			}
			c.gmp()
			c.ggx()

			assert.Equal(t, p.exp, c.gx)
			assert.Equal(t, p.k, c.k)
		}
		t.Run(n, f)
	}

}
