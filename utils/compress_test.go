package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var rleTestData = map[string]struct {
	s      []byte
	expect []byte
}{
	"case1": {
		s:      []byte{'A', 'A', 'A', 'B', 'B', 'C', 'C'},
		expect: []byte{0x83, 'A', 0x04, 'B', 'B', 'C', 'C'},
	},
	"case2": {
		s: []byte{
			'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a',
			'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a',
		},
		expect: []byte{0x9b, 'a'},
	},
	"case3": {
		s: []byte{
			'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a',
			'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a',
			'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a',
			'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a',
			'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a',
			'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a',
			'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a',
			'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a',
			'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a',
			'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a',
		},
		expect: []byte{0xff, 'a', 0xa1, 'a'},
	},
	"case4": {
		s: []byte{
			'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h',
			'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h',
			'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h',
			'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h',
			'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h',
			'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h',
			'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h',
			'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h',
			'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h',
			'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h',
		},
		expect: []byte{
			0x7f,
			'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h',
			'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h',
			'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h',
			'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h',
			'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h',
			'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h',
			'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h',
			'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'a', 'b', 'c', 'd', 'e', 'f', 'g',
			0x21, 'h',
			'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h',
			'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h',
		},
	},
	"case5": {
		s: []byte{
			0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
		},
		expect: []byte{0x10, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
	},
}

func Test_rle(t *testing.T) {
	var rle, err = NewRLE()
	if err != nil {
		t.Fatal(err)
		return
	}
	for n, p := range rleTestData {
		f := func(t *testing.T) {
			var cc []byte
			cc, err = rle.Compress(p.s)
			if err != nil {
				t.Fatal(err)
				return
			}
			assert.Equal(t, p.expect, cc)
			p.expect, err = rle.Uncompress(cc)
			if err != nil {
				t.Fatal(err)
				return
			}
			assert.Equal(t, p.s, p.expect)
		}
		t.Run(n, f)
	}
}

var rleRandTestData = map[string]struct {
	length int
}{
	"case1": {
		length: 512,
	},
	"case2": {
		length: 1024,
	},
	"case3": {
		length: 2048,
	},
	"case4": {
		length: 4096,
	},
	"case5": {
		length: 8192,
	},
	"case6": {
		length: 16384,
	},
	"case7": {
		length: 65536,
	},
	"case8": {
		length: 10240,
	},
	"case9": {
		length: 20480,
	},
}

func Test_rle_random(t *testing.T) {
	var rle, err = NewRLE()
	if err != nil {
		t.Fatal(err)
		return
	}
	for n, p := range rleRandTestData {
		f := func(t *testing.T) {
			var data = RandsString(p.length)
			var cc []byte
			cc, err = rle.Compress([]byte(data))
			if err != nil {
				t.Fatal(err)
				return
			}
			cc, err = rle.Uncompress(cc)
			if err != nil {
				t.Fatal(err)
				return
			}
			assert.Equal(t, []byte(data), cc)
		}
		t.Run(n, f)
	}
}
