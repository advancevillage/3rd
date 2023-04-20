package cc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_G(t *testing.T) {
	var data = map[string]struct {
		n  int
		gx []byte
		G  [][]byte
		H  [][]byte
	}{
		"(7,4)|gx=x³+x+1": {
			n:  7,
			gx: []byte{1, 1, 0, 1},
			G: [][]byte{
				{1, 0, 1, 0, 0, 0, 1},
				{1, 1, 1, 0, 0, 1, 0},
				{0, 1, 1, 0, 1, 0, 0},
				{1, 1, 0, 1, 0, 0, 0},
			},
			H: [][]byte{
				{0, 0, 1, 0, 1, 1, 1},
				{0, 1, 0, 1, 1, 1, 0},
				{1, 0, 0, 1, 0, 1, 1},
			},
		},
		"(8,5)|gx=x³+x²+x+1": {
			n:  8,
			gx: []byte{1, 1, 1, 1},
			G: [][]byte{
				{1, 1, 1, 0, 0, 0, 0, 1},
				{0, 0, 1, 0, 0, 0, 1, 0},
				{0, 1, 0, 0, 0, 1, 0, 0},
				{1, 0, 0, 0, 1, 0, 0, 0},
				{1, 1, 1, 1, 0, 0, 0, 0},
			},
			H: [][]byte{
				{0, 0, 1, 1, 0, 0, 1, 1},
				{0, 1, 0, 1, 0, 1, 0, 1},
				{1, 0, 0, 1, 1, 0, 0, 1},
			},
		},
	}

	for n, p := range data {
		f := func(t *testing.T) {
			c, err := NewWithGx(p.n, p.gx)
			if err != nil {
				t.Fatal(err)
				return
			}
			assert.Equal(t, p.G, c.G())
			assert.Equal(t, p.H, c.H())
		}
		t.Run(n, f)
	}

}

func Test_Encode(t *testing.T) {
	var data = map[string]struct {
		n   int
		gx  []byte
		m   []byte
		exp []byte
	}{
		"(7,4)|gx=x³+x+1": {
			n:   7,
			gx:  []byte{1, 1, 0, 1},
			m:   []byte{1, 1, 0, 1},
			exp: []byte{0, 0, 0, 1, 1, 0, 1},
		},
		"(8,5)|gx=x³+x²+x+1": {
			n:   8,
			gx:  []byte{1, 1, 1, 1},
			m:   []byte{1, 1, 0, 1, 1},
			exp: []byte{1, 0, 1, 1, 1, 0, 1, 1},
		},
	}

	for n, p := range data {
		f := func(t *testing.T) {
			c, err := NewWithGx(p.n, p.gx)
			if err != nil {
				t.Fatal(err)
				return
			}
			act, err := c.(cc).encode(p.m)
			if err != nil {
				t.Fatal(err)
				return
			}
			assert.Equal(t, p.exp, act)
		}
		t.Run(n, f)
	}

}
