//author: richard
package utils

import (
	"encoding/binary"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
)

func UUID() string {
	return uuid.New().String()
}

func UuId() []byte {
	return []byte(UUID())
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
