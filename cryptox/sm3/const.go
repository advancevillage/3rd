package sm3

// GB/T 32905-2016
// The size of an SM3 checksum in bytes.
// link: http://c.gb688.cn/bzgk/gb/showGb?type=online&hcno=45B1A67F20F3BF339211C391E9278F5E
const Size = 32

const BlockSize = 64

const (
	init0 = 0x7380166f
	init1 = 0x4914b2b9
	init2 = 0x172442d7
	init3 = 0xda8a0600
	init4 = 0xa96f30bc
	init5 = 0x163138aa
	init6 = 0xe38dee4d
	init7 = 0xb0fb0e4e
)

func kk(i int) uint32 {
	if i < 16 {
		return 0x79cc4519
	} else {
		return 0x7a879d8a
	}
}

func ff(i int, x, y, z uint32) uint32 {
	if i < 16 {
		return x ^ y ^ z
	} else {
		return (x & y) | (x & z) | (y & z)
	}
}

func gg(i int, x, y, z uint32) uint32 {
	if i < 16 {
		return x ^ y ^ z
	} else {
		return (x & y) | (^x & z)
	}
}

func p0(x uint32) uint32 {
	return x ^ rol32(x, 9) ^ rol32(x, 17)
}

func p1(x uint32) uint32 {
	return x ^ rol32(x, 15) ^ rol32(x, 23)
}

func rol32(x uint32, n int) uint32 {
	return (x << n) | ((x & 0xffffffff) >> (32 - n))
}
