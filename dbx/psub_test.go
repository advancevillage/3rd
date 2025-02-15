package dbx_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/advancevillage/3rd/dbx"
	"github.com/advancevillage/3rd/logx"
	"github.com/advancevillage/3rd/mathx"
	"github.com/stretchr/testify/assert"
)

var _ dbx.Subscriber = (*testPubSub)(nil)

type testPubSub struct {
	payload string
	logger  logx.ILogger
	tt      *testing.T
}

func (t *testPubSub) Subscribe(ctx context.Context, payload string) error {
	t.logger.Infow(ctx, "payload", "send", t.payload, "recv", payload)
	assert.Equal(t.tt, t.payload, payload)
	return nil
}

// 倒序查看:
//
//	xrevrange ${topic} + - count 2
//
// 查看消费组:
//
//	xinfo groups ${topic}
func Test_pubsub(t *testing.T) {
	logger, err := logx.NewLogger("debug")
	if err != nil {
		t.Fatal(err)
		return
	}

	var data = map[string]struct {
		ps       *testPubSub
		topic    string
		cg       string
		customer string
	}{
		"case1": {
			ps:       &testPubSub{payload: mathx.RandStr(10), logger: logger, tt: t},
			cg:       "cg-g-1",
			topic:    "pub-c-1",
			customer: fmt.Sprintf("cg-c-%d", 1),
		},
		"case2": {
			ps:       &testPubSub{payload: mathx.RandStr(10), logger: logger, tt: t},
			cg:       mathx.RandStr(10),
			topic:    mathx.RandStr(10),
			customer: fmt.Sprintf("cg-c-%d", mathx.GId()),
		},
	}

	for n, v := range data {
		f := func(t *testing.T) {
			ctx := context.WithValue(context.TODO(), logx.TraceId, mathx.UUID())
			ctx, cancel := context.WithDeadline(ctx, time.Now().Add(time.Second))
			defer cancel()

			p, err := dbx.NewProducer(ctx, logger, dbx.WithPubSubChannel(v.topic), dbx.WithConsumerGroup(v.cg))
			if err != nil {
				t.Fatal(err)
				return
			}
			err = dbx.NewConsumer(ctx, logger, v.customer, dbx.WithPubSubChannel(v.topic), dbx.WithConsumerGroup(v.cg), dbx.WithSubscriber(v.ps))
			if err != nil {
				t.Fatal(err)
				return
			}
			err = p.Publish(ctx, v.ps.payload)
			if err != nil {
				t.Fatal(err)
				return
			}

			<-ctx.Done()
			time.Sleep(time.Second)
		}
		t.Run(n, f)
	}
}

func Test_deay(t *testing.T) {
	logger, err := logx.NewLogger("debug")
	assert.Nil(t, err)

	var data = map[string]struct {
		ps       *testPubSub
		topic    string
		cg       string
		customer string
		delay    time.Duration
	}{
		"case1": {
			ps:       &testPubSub{payload: mathx.RandStr(10), logger: logger, tt: t},
			cg:       "cg-g-1-d",
			topic:    "pub-c-1-d",
			customer: fmt.Sprintf("cg-c-%d", 1),
			delay:    time.Second * 3,
		},
		"case2": {
			ps:       &testPubSub{payload: mathx.RandStr(10), logger: logger, tt: t},
			cg:       mathx.RandStr(10),
			topic:    mathx.RandStr(10),
			customer: fmt.Sprintf("cg-c-%d", mathx.GId()),
			delay:    0,
		},
	}
	for n, v := range data {
		f := func(t *testing.T) {
			ctx := context.WithValue(context.TODO(), logx.TraceId, mathx.UUID())
			ctx, cancel := context.WithDeadline(ctx, time.Now().Add(time.Second*5))
			defer cancel()

			p, err := dbx.NewProducer(ctx, logger, dbx.WithPubSubChannel(v.topic), dbx.WithConsumerGroup(v.cg))
			if err != nil {
				t.Fatal(err)
				return
			}
			err = dbx.NewConsumer(ctx, logger, v.customer, dbx.WithPubSubChannel(v.topic), dbx.WithConsumerGroup(v.cg), dbx.WithSubscriber(v.ps))
			if err != nil {
				t.Fatal(err)
				return
			}
			err = p.Delay(ctx, v.ps.payload, v.delay)
			if err != nil {
				t.Fatal(err)
				return
			}

			<-ctx.Done()
			time.Sleep(time.Second)
		}
		t.Run(n, f)
	}
}
