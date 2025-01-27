package dbx

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/advancevillage/3rd/logx"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
)

type Publisher interface {
	Publish(ctx context.Context, payload string) error
}

type Subscriber interface {
	Subscribe(ctx context.Context, payload string) error
}

type PubSubOption interface {
	apply(*pubsubOption)
}

func WithPublisherQL(ql int64) PubSubOption {
	return newFuncPubSubOption(func(o *pubsubOption) {
		o.ql = ql
	})
}

func WithPubSubDNS(dns string) PubSubOption {
	return newFuncPubSubOption(func(o *pubsubOption) {
		o.dns = dns
	})
}

func WithPubSubChannel(channel string) PubSubOption {
	return newFuncPubSubOption(func(o *pubsubOption) {
		o.channel = channel
	})
}

func WithConsumerGroup(cg string) PubSubOption {
	return newFuncPubSubOption(func(o *pubsubOption) {
		o.cg = cg
	})
}

func WithSubscriber(sub Subscriber) PubSubOption {
	return newFuncPubSubOption(func(o *pubsubOption) {
		o.subscriber = sub
	})
}

type pubsubOption struct {
	cg         string     // 消费组 consumer group
	ql         int64      // 生产者 最大消息长度
	dns        string     // 组件地址
	channel    string     // 订阅主题频道
	subscriber Subscriber // 消费组中的订阅者
}

var defaultPubSubOptions = pubsubOption{
	cg:         "advancevillage",
	ql:         100,
	dns:        "redis://127.0.0.1:6379/0?protocol=3",
	channel:    "advancevillage",
	subscriber: &emptyConsumer{},
}

type funcPubSubOption struct {
	f func(*pubsubOption)
}

func (fdo *funcPubSubOption) apply(do *pubsubOption) {
	fdo.f(do)
}

func newFuncPubSubOption(f func(*pubsubOption)) *funcPubSubOption {
	return &funcPubSubOption{
		f: f,
	}
}

var _ Publisher = (*producerRedis)(nil)

type producerRedis struct {
	opts   pubsubOption
	rdb    *redis.Client
	logger logx.ILogger
}

const (
	X_TICKET_TIME    = "x-ticket-time"
	X_TICKET_PAYLOAD = "x-ticket-payload"
)

func NewProducer(ctx context.Context, logger logx.ILogger, opt ...PubSubOption) (Publisher, error) {
	return newProducerRedis(ctx, logger, opt...)
}

func newProducerRedis(ctx context.Context, logger logx.ILogger, opt ...PubSubOption) (*producerRedis, error) {
	// 1. 设置配置
	opts := defaultPubSubOptions
	for _, o := range opt {
		o.apply(&opts)
	}
	// 2. 创建连接
	rdbOpts, err := redis.ParseURL(opts.dns)
	if err != nil {
		logger.Errorw(ctx, "redis parse url failed", "err", err, "dns", opts.dns)
		return nil, err
	}
	// 3. 返回对象
	return &producerRedis{
		opts:   opts,
		rdb:    redis.NewClient(rdbOpts),
		logger: logger,
	}, nil
}

func (p *producerRedis) Publish(ctx context.Context, payload string) error {
	// Redis Stream是惰性创建的，第一次执行XADD或XGROUP时才会真正创
	err := p.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: p.opts.channel,
		MaxLen: p.opts.ql,
		Values: map[string]interface{}{
			logx.TraceId:     ctx.Value(logx.TraceId),
			X_TICKET_TIME:    time.Now().UnixNano() / 1e6,
			X_TICKET_PAYLOAD: payload,
		},
	}).Err()
	if err != nil {
		p.logger.Errorw(ctx, "redis publish failed", "err", err)
	}
	return err
}

type consumerRedis struct {
	opts       pubsubOption
	rdb        *redis.Client
	logger     logx.ILogger
	consumerId string
}

func NewConsumer(ctx context.Context, logger logx.ILogger, consumerId string, opt ...PubSubOption) error {
	c, err := newConsumerRedis(ctx, logger, consumerId, opt...)
	if err != nil {
		return err
	}
	c.logger.Infow(ctx, "consumer start subscribe")
	return nil
}

func newConsumerRedis(ctx context.Context, logger logx.ILogger, consumerId string, opt ...PubSubOption) (*consumerRedis, error) {
	// 1. 设置配置
	opts := defaultPubSubOptions
	for _, o := range opt {
		o.apply(&opts)
	}
	// 2. 创建连接
	rdbOpts, err := redis.ParseURL(opts.dns)
	if err != nil {
		logger.Errorw(ctx, "redis parse url failed", "err", err, "dns", opts.dns)
		return nil, err
	}
	// 3. 设置对象
	c := &consumerRedis{
		opts:       opts,
		rdb:        redis.NewClient(rdbOpts),
		logger:     logger,
		consumerId: consumerId,
	}

	// 4. 创建消费组
	err = c.createGroupIfNotExists(ctx)
	if err != nil {
		return nil, err
	}

	// 5. 订阅
	go c.subscribe(ctx)

	return c, nil
}

func (c *consumerRedis) createGroupIfNotExists(ctx context.Context) error {
	// 1. 查询消费组
	groups, err := c.rdb.XInfoGroups(ctx, c.opts.channel).Result()
	if err != nil && !strings.Contains(err.Error(), "ERR no such key") {
		return err
	}

	for i := range groups {
		if c.opts.cg == groups[i].Name {
			return nil
		}
	}
	// 2. 创建消费组(从头开始消费)
	err = c.rdb.XGroupCreateMkStream(ctx, c.opts.channel, c.opts.cg, "0").Err()
	if err != nil {
		c.logger.Errorw(ctx, "redis create group failed", "err", err, "channel", c.opts.channel, "cg", c.opts.cg)
		return err
	}
	c.logger.Infow(ctx, "redis create group success", "channel", c.opts.channel, "cg", c.opts.cg)
	return nil
}

func (c *consumerRedis) subscribe(ctx context.Context) {
	var (
		g  = new(errgroup.Group)
		ch = make(chan struct{}, 3)
	)
	for {
		select {
		case <-ctx.Done():
			goto exit

		default:
			entries, err := c.rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
				Group:    c.opts.cg,
				Consumer: c.consumerId,
				Streams:  []string{c.opts.channel, ">"},
				Count:    5,
				Block:    0,
				NoAck:    false,
			}).Result()
			if err != nil {
				c.logger.Errorw(ctx, "redis read group failed", "err", err)
				continue
			}

			for i := range entries {
				entry := entries[i]

				for j := range entry.Messages {
					msg := entry.Messages[j]

					var (
						msgId  = msg.ID
						values = msg.Values
					)
					// 1. 提取traceId
					sctx := context.WithValue(context.Background(), logx.TraceId, values[logx.TraceId])

					// 2. 提取时间
					stime, err := strconv.ParseInt(fmt.Sprint(values[X_TICKET_TIME]), 10, 64)
					if err != nil {
						c.logger.Errorw(sctx, "subscribe parse X_TICKET_TIME failed", "msgId", msgId, "err", err, "stime", values[X_TICKET_TIME])
						continue
					}
					etime := time.Now().UnixNano() / 1e6

					// 3. 提取内容
					payload := fmt.Sprint(values[X_TICKET_PAYLOAD])
					if len(payload) <= 0 {
						c.logger.Infow(ctx, "subscribe parse payload empty", "msgId", msgId, "stime", stime, "etime", etime, "delay", etime-stime)
						continue
					}

					// 4. 处理数据
					ch <- struct{}{}
					g.Go(func() error {
						defer func() {
							<-ch
						}()
						c.logger.Infow(sctx, "subscribe receive payload", "msgId", msgId, "delay", etime-stime)
						err := c.opts.subscriber.Subscribe(sctx, payload)
						if err != nil {
							c.logger.Errorw(sctx, "subscribe handle payload failed", "msgId", msgId, "err", err)
							return nil
						}
						err = c.rdb.XAck(sctx, c.opts.channel, c.opts.cg, msgId).Err()
						if err != nil {
							c.logger.Errorw(sctx, "subscribe ack msg failed", "msgId", msgId, "err", err)
							return nil
						}
						err = c.rdb.XDel(sctx, c.opts.channel, msgId).Err()
						if err != nil {
							c.logger.Errorw(sctx, "subscribe clear msg failed", "msgId", msgId, "err", err)
							return nil
						}
						c.logger.Infow(sctx, "subscribe handle msg success", "msgId", msgId)
						return nil
					})
				}
			}

		}
	}

exit:
	if err := g.Wait(); err != nil {
		c.logger.Errorw(ctx, "subscribe handle payload failed", "err", err)
	}
	c.logger.Infow(ctx, "subscribe exit", "time", time.Now().UnixNano()/1e6)
}

var _ Subscriber = (*emptyConsumer)(nil)

type emptyConsumer struct {
}

func (e *emptyConsumer) Subscribe(ctx context.Context, payload string) error {
	return nil
}
