package netx

import (
	"time"

	"github.com/advancevillage/3rd/database"
	"github.com/advancevillage/3rd/ecies"
)

const (
	nBucket     = 256
	bucketLimit = 20
)

type IDHT interface {
}

type kad struct {
	db      database.ILevelDB
	uc      IUDPClient
	buckets [nBucket]*bucket
	nursery []ecies.IENode
}

//@overview: kad table
//@author: richard.sun
//@param:
//1. 核心是确定xor索引方式
type bucket struct {
	nodes []*node
}

type node struct {
	id       ecies.IENode
	ts       time.Time
	liveness uint
}

func NewDHT() (IDHT, error) {

	return nil, nil
}
