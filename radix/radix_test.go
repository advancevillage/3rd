package radix

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

var u32Test = map[string]struct {
	smpl int
}{
	"case1": {
		smpl: 10000,
	},
	"case2": {
		smpl: 100000,
	},
	"case3": {
		smpl: 1000000,
	},
	"case4": {
		smpl: 200,
	},
}

type u32T struct {
	key   uint32
	mask  uint32
	value uint32
	err   error
}

type u64T struct {
	key   uint64
	mask  uint64
	value uint64
	err   error
}

func Test_u32_radix(t *testing.T) {
	for n, p := range u32Test {
		f := func(t *testing.T) {
			var r = NewRadixTree()

			data := genU32T(p.smpl)

			for _, v := range data {
				err := r.AddU32(v.key, v.mask, v.value)
				if err != nil {
					t.Fatal(err)
					return
				}
			}

			data2 := r.ListU32()

			var m1 = make(map[uint32]bool)
			var m2 = make(map[uint32]bool)

			for _, v := range data {
				m1[v.value] = true
			}
			for _, v := range data2 {
				m2[v] = true
			}
			assert.Equal(t, m1, m2)

			data = append(data, &u32T{
				key:   0x01ffffff,
				value: 0x01,
				mask:  0xffffffff,
				err:   notFound,
			})

			for _, v := range data {
				val, err := r.GetU32(v.key)
				if err != nil {
					assert.Equal(t, v.err, err)
				} else {
					assert.Equal(t, v.value, val)
				}
			}

			val := data[rand.Intn(p.smpl)]
			val.err = notFound

			err := r.DelU32(val.key, val.mask)
			if err != nil {
				t.Fatal(err)
				return
			}

			val2, err := r.GetU32(val.key)
			if err != nil {
				assert.Equal(t, val.err, err)
			} else {
				assert.Equal(t, val.value, val2)
			}
		}
		t.Run(n, f)
	}
}

func genU32T(n int) []*u32T {
	var m = make(map[uint32]*u32T)

	for n > len(m) {
		var a = uint8(rand.Intn(254) + 1)
		var b = uint8(rand.Intn(254) + 1)
		var c = uint8(rand.Intn(254) + 1)
		var d = uint8(rand.Intn(254) + 1)

		var k = uint32(a)<<24 | uint32(b)<<16 | uint32(c)<<8
		if _, ok := m[k]; ok {
			continue
		} else {
			v := uint32(rand.Int31())
			m[k] = &u32T{
				key:   k | uint32(d),
				value: v,
				mask:  0xffffff00,
				err:   nil,
			}
		}
	}

	var d []*u32T

	for _, v := range m {
		d = append(d, v)
	}

	return d
}

var u64Test = map[string]struct {
	smpl int
}{
	"case1": {
		smpl: 10000,
	},
	"case2": {
		smpl: 100000,
	},
	"case3": {
		smpl: 1000000,
	},
	"case4": {
		smpl: 200,
	},
}

func Test_u64_radix(t *testing.T) {
	for n, p := range u64Test {
		f := func(t *testing.T) {
			var r = NewRadixTree()

			data := genU64T(p.smpl)
			for _, v := range data {
				err := r.AddU64(v.key, v.mask, v.value)
				if err != nil {
					t.Fatal(err)
					return
				}
			}

			data2 := r.ListU64()

			var m1 = make(map[uint64]bool)
			var m2 = make(map[uint64]bool)

			for _, v := range data {
				m1[v.value] = true
			}
			for _, v := range data2 {
				m2[v] = true
			}
			assert.Equal(t, m1, m2)

			data = append(data, &u64T{
				key:   0x01ffffffffffff00,
				value: 0x01,
				mask:  0xffffffffffffffff,
				err:   notFound,
			})

			for _, v := range data {
				val, err := r.GetU64(v.key)
				if err != nil {
					assert.Equal(t, v.err, err)
				} else {
					assert.Equal(t, v.value, val)
				}
			}

			val := data[rand.Intn(p.smpl)]
			val.err = notFound

			err := r.DelU64(val.key, val.mask)
			if err != nil {
				t.Fatal(err)
				return
			}

			val2, err := r.GetU64(val.key)
			if err != nil {
				assert.Equal(t, val.err, err)
			} else {
				assert.Equal(t, val.value, val2)
			}
		}
		t.Run(n, f)
	}
}

func genU64T(n int) []*u64T {
	var m = make(map[uint64]*u64T)

	for n > len(m) {
		var a = uint8(rand.Intn(254) + 1)
		var b = uint8(rand.Intn(254) + 1)
		var c = uint8(rand.Intn(254) + 1)
		var d = uint8(rand.Intn(254) + 1)
		var e = uint8(rand.Intn(254) + 1)
		var f = uint8(rand.Intn(254) + 1)
		var g = uint8(rand.Intn(254) + 1)
		var h = uint8(rand.Intn(254) + 1)

		var k = uint64(a)<<56 | uint64(b)<<48 | uint64(c)<<40 | uint64(d)<<32 | uint64(e)<<24 | uint64(f)<<16
		if _, ok := m[k]; ok {
			continue
		} else {
			m[k] = &u64T{
				key:   k | uint64(g)<<8 | uint64(h),
				value: k | uint64(g)<<8 | uint64(h),
				mask:  0xffffffffffff0000,
				err:   nil,
			}
		}
	}

	var d []*u64T

	for _, v := range m {
		d = append(d, v)
	}

	return d
}
