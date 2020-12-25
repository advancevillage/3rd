//author: richard
package utils

import (
	"encoding/binary"
	"math/rand"
	"strconv"
	"sync"
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
	u[9] = u[0]>>6&0x1F | (u[7]<<2)&0x1F
	u[10] = u[1] >> 1 & 0x1F
	u[11] = u[2]>>4&0x1F | (u[5]<<4)&0x1F
	u[12] = u[3]>>7 | (u[9]<<1)&0x1F
	u[13] = u[4] >> 2 & 0x1F
	u[14] = u[5]>>5 | (u[3]<<3)&0x1F
	u[15] = u[9] & 0x1F
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

/////////SnowFlake
//@link: https://segmentfault.com/a/1190000014767902
//@link: https://www.dazhuanlan.com/2019/11/04/5dbfbb901bef2/

const (
	numberBits      uint8 = 12                      // 每个集群的节点生成的ID数最大位数
	workerBits      uint8 = 10                      // 工作机器的ID位数
	numberMax       int64 = -1 ^ (-1 << numberBits) // ID序号的最大值  4096
	workerIdMax     int64 = -1 ^ (-1 << workerBits) // 工作机器的ID最大值 1024
	timeShift             = workerBits + numberBits // 时间戳向左偏移量
	workerShift           = numberBits              // 节点ID向左偏移数
	sub             int64 = 1525705533000           // 减去现在的时间戳
	defaultWorkerId       = 1                       // 默认worker
)

var sf = &snowFlake{timestamp: 0, workerId: defaultWorkerId, number: 0}

type snowFlake struct {
	mu        sync.RWMutex
	timestamp int64 // 上一次生成ID的时间戳
	workerId  int64 // 节点ID
	number    int64 // 已经生成的ID数
}

//生成全局唯一ID
func SnowFlakeId() int64 {
	sf.mu.Lock()
	defer sf.mu.Unlock()
	now := time.Now().UnixNano() / 1e6
	if sf.timestamp == now {
		sf.number++
		// 判断是否已经超出最大的限制的ID
		if sf.number > numberMax {
			// 等待下一毫秒
			for now <= sf.timestamp {
				now = time.Now().UnixNano() / 1e6
			}
		}
	} else {
		// 新的一毫秒将number 修改为0 timestamp修改为now
		sf.number = 0
		sf.timestamp = now
	}
	id := (now-sub)<<timeShift | sf.workerId<<workerShift | sf.number
	return id
}

func SnowFlakeIdString() string {
	return strconv.FormatInt(SnowFlakeId(), 10)
}

func SnowFlakeIdBytes(n int) []byte {
	var b = make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(SnowFlakeId()))
	if n <= 8 {
		return b[:n]
	}
	var bn = make([]byte, n)
	for i := 0; i < n; i++ {
		bn[i] = b[rand.Intn(8)]
	}
	return bn
}
