package mathx

import (
	"sync"

	"github.com/bwmarrin/snowflake"
	"github.com/google/uuid"
)

type IDGenerator interface {
	Generate() int64
}

var _ IDGenerator = (*sf)(nil)

type sf struct {
	n *snowflake.Node
}

func NewSnowFlake(seed int64) (IDGenerator, error) {
	return newSf(seed)
}

func newSf(seed int64) (*sf, error) {
	n, err := snowflake.NewNode(seed)
	if err != nil {
		return nil, err
	}
	return &sf{n: n}, nil
}

func (g *sf) Generate() int64 {
	return g.n.Generate().Int64()
}

var (
	gid  IDGenerator
	once sync.Once
)

func init() {
	once.Do(func() {
		gid, _ = NewSnowFlake(1)
	})
}

func GId() int64 {
	return gid.Generate()
}

func UUID() string {
	return uuid.New().String()
}
