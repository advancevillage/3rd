package sm3

import (
	"encoding/binary"
	"hash"
)

func New() hash.Hash {
	d := new(digest)
	d.Reset()
	return d
}

// GB/T 32905-2016
type digest struct {
	s  [8]uint32
	x  [BlockSize]byte
	nx int
	l  uint64
}

func (d *digest) Reset() {
	d.s[0] = init0
	d.s[1] = init1
	d.s[2] = init2
	d.s[3] = init3
	d.s[4] = init4
	d.s[5] = init5
	d.s[6] = init6
	d.s[7] = init7
	d.nx = 0
	d.l = 0
}

func (d *digest) Size() int { return Size }

func (d *digest) BlockSize() int { return BlockSize }

func (d *digest) Write(p []byte) (nn int, err error) {

	nn = len(p)

	d.l += uint64(nn)

	if d.nx > 0 {
		n := copy(d.x[d.nx:], p)
		d.nx += n
		if d.nx == BlockSize {
			blockGeneric(d, d.x[:])
			d.nx = 0
		}
		p = p[n:]
	}

	if len(p) >= BlockSize {
		n := len(p) &^ (BlockSize - 1)
		blockGeneric(d, p[:n])
		p = p[n:]
	}

	if len(p) > 0 {
		d.nx = copy(d.x[:], p)
	}

	return
}

func (d *digest) Sum(in []byte) []byte {
	d0 := *d
	hash := d0.checkSum()
	return append(in, hash[:]...)
}

func (d *digest) checkSum() [Size]byte {
	// Append 0x80 to the end of the message and then append zeros
	// until the length is a multiple of 56 bytes. Finally append
	// 8 bytes representing the message length in bits.
	//
	// 1 byte end marker :: 0-63 padding bytes :: 8 byte length
	tmp := [1 + 63 + 8]byte{0x80}
	pad := (55 - d.l) % 64                          // calculate number of padding bytes
	binary.BigEndian.PutUint64(tmp[1+pad:], d.l<<3) // append length in bits
	d.Write(tmp[:1+pad+8])

	// The previous write ensures that a whole number of
	// blocks (i.e. a multiple of 64 bytes) have been hashed.
	if d.nx != 0 {
		panic("d.nx != 0")
	}

	var digest [Size]byte
	binary.BigEndian.PutUint32(digest[0x00:], d.s[0])
	binary.BigEndian.PutUint32(digest[0x04:], d.s[1])
	binary.BigEndian.PutUint32(digest[0x08:], d.s[2])
	binary.BigEndian.PutUint32(digest[0x0c:], d.s[3])
	binary.BigEndian.PutUint32(digest[0x10:], d.s[4])
	binary.BigEndian.PutUint32(digest[0x14:], d.s[5])
	binary.BigEndian.PutUint32(digest[0x18:], d.s[6])
	binary.BigEndian.PutUint32(digest[0x1c:], d.s[7])
	return digest
}

func Sum(data []byte) [Size]byte {
	var d digest
	d.Reset()
	d.Write(data)
	return d.checkSum()
}
