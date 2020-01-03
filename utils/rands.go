//author: richard
package utils

import (
	"math/rand"
	"time"
	"unsafe"
)

const (
	letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numberBytes = "0123456789"
)

const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

//@brief: 高效随机函数
//@link: https://www.flysnow.org/2019/09/30/how-to-generate-a-random-string-of-a-fixed-length-in-go.html#benchmark-%E6%80%A7%E8%83%BD%E6%B5%8B%E8%AF%95
func rands(n int, letters string) string {
	b := make([]byte, n)
	src := rand.NewSource(time.Now().UnixNano())
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letters) {
			b[i] = letters[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	return *(*string)(unsafe.Pointer(&b))
}

func RandsString(n int) string {
	return rands(n, letterBytes)
}

func RandsInt(up int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(up)
}

func RandsNumberString(n int) string {
	return rands(n, numberBytes)
}
