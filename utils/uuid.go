//author: richard
package utils

import (
	"sync/atomic"
	"time"
)

type uuid [16]byte

//@link: https://m.ancii.com/ah7dpzzl/
func generator(a time.Time) uuid {
	var u uuid
	var seq uint32
	var hardware []byte
	var base = time.Date(1994, time.February, 15, 0, 0, 0, 0, time.UTC).Unix()
	utc := a.In(time.UTC)
	t := uint64(utc.Unix()-base)*10000000 + uint64(utc.Nanosecond()/100)
	u[0], u[1], u[2], u[3] = byte(t>>24), byte(t>>16), byte(t>>8), byte(t)
	u[4], u[5] = byte(t>>40), byte(t>>32)
	u[6], u[7] = byte(t>>56)&0x0F, byte(t>>48)
	clock := atomic.AddUint32(&seq, 1)
	u[8] = byte(clock >> 8)
	u[9] = byte(clock)
	copy(u[10:], hardware)
	u[6] |= 0x10 // set version to 1 (time based uuid)
	u[8] &= 0x3F // clear variant
	u[8] |= 0x80 // set to IETF variant
	u[9]  = u[8]>>6&0x1F|(u[7]<<2)&0x1F
	u[10] = u[9]>>1&0x1F
	u[11] = u[10]>>4&0x1F|(u[8]<<4)&0x1F
	u[12] = u[11]>>7|(u[9]<<1)&0x1F
	u[13] = u[12]>>2&0x1F
	u[14] = u[13]>>5|(u[10]<<3)&0x1F
	u[15] = u[14]&0x1F
	return u
}

func (u *uuid) toString() string {
	var offsets = [...]int{0, 2, 4, 6, 9, 11, 14, 16, 19, 21, 24, 26, 28, 30, 32, 34}
	const hexString = "0123456789abcdef"
	r := make([]byte, 36)
	for i, b := range u {
		r[offsets[i]] = hexString[b>>4]
		r[offsets[i]+1] = hexString[b&0xF]
	}
	r[8] = '-'
	r[13] = '-'
	r[18] = '-'
	r[23] = '-'
	return string(r)
}

func UUID() string {
	u := generator(time.Now())
	return u.toString()
}