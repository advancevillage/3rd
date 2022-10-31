package sm4

import (
	"encoding/binary"
	"unsafe"
)

// GB/T 32907-2016
// http://c.gb688.cn/bzgk/gb/showGb?type=online&hcno=7803DE42D3BC5E80B0C3E5D8E873D56A
func encryptBlockGo(xk []uint32, dst, src []byte) {
	_ = src[15] // early bounds check
	s0 := binary.BigEndian.Uint32(src[0:4])
	s1 := binary.BigEndian.Uint32(src[4:8])
	s2 := binary.BigEndian.Uint32(src[8:12])
	s3 := binary.BigEndian.Uint32(src[12:16])

	for i := 0; i < 32; i++ {

		si := F(s0, s1, s2, s3, xk[i])

		s0 = s1
		s1 = s2
		s2 = s3
		s3 = si
	}

	_ = dst[15] // early bounds check
	binary.BigEndian.PutUint32(dst[0:4], s3)
	binary.BigEndian.PutUint32(dst[4:8], s2)
	binary.BigEndian.PutUint32(dst[8:12], s1)
	binary.BigEndian.PutUint32(dst[12:16], s0)
}

func expandEncKeyGo(key []byte) []uint32 {
	_ = key[15] // early bounds check
	s0 := binary.BigEndian.Uint32(key[0:4])
	s1 := binary.BigEndian.Uint32(key[4:8])
	s2 := binary.BigEndian.Uint32(key[8:12])
	s3 := binary.BigEndian.Uint32(key[12:16])

	k0 := s0 ^ fk0
	k1 := s1 ^ fk1
	k2 := s2 ^ fk2
	k3 := s3 ^ fk3

	rk := make([]uint32, 32)

	for i := 0; i < 32; i++ {
		x := k0 ^ T2(k1^k2^k3^ck[i])

		k0 = k1
		k1 = k2
		k2 = k3
		k3 = x

		rk[i] = x
	}
	return rk
}

func expandDecKeyGo(key []byte) []uint32 {
	_ = key[15] // early bounds check
	s0 := binary.BigEndian.Uint32(key[0:4])
	s1 := binary.BigEndian.Uint32(key[4:8])
	s2 := binary.BigEndian.Uint32(key[8:12])
	s3 := binary.BigEndian.Uint32(key[12:16])

	k0 := s0 ^ fk0
	k1 := s1 ^ fk1
	k2 := s2 ^ fk2
	k3 := s3 ^ fk3

	rk := make([]uint32, 32)

	for i := 0; i < 32; i++ {
		x := k0 ^ T2(k1^k2^k3^ck[i])

		k0 = k1
		k1 = k2
		k2 = k3
		k3 = x

		rk[31-i] = x
	}

	return rk
}

// copy from https://github.com/golang/go/blob/15da892a4950a4caac987ee72c632436329f62d5/src/crypto/internal/subtle/aliasing.go#L30
func inexactOverlap(x, y []byte) bool {
	if len(x) == 0 || len(y) == 0 || &x[0] == &y[0] {
		return false
	}
	return anyOverlap(x, y)
}

func anyOverlap(x, y []byte) bool {
	return len(x) > 0 && len(y) > 0 &&
		uintptr(unsafe.Pointer(&x[0])) <= uintptr(unsafe.Pointer(&y[len(y)-1])) &&
		uintptr(unsafe.Pointer(&y[0])) <= uintptr(unsafe.Pointer(&x[len(x)-1]))
}
