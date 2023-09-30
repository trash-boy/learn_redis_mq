package learn_redis_mq

import (
	"context"
	"errors"
	"learn_redis_mq/log"
	"learn_redis_mq/redis"
)

type MsgCallback func(ctx context.Context, msg *redis.MsgEntity)error

type Consumer struct {
	ctx context.Context
	stop context.CancelFunc

	callbackFunc MsgCallback
	client *redis.Client

	topic string
	groupID string
	consumerID string

	failureCnts map[redis.MsgEntity]int

	opts *ConsumerOptions
}

func (c *Consumer) checkParam() error {
	if c.callbackFunc == nil{
		return errors.New("callback function can not be empty")
	}
	if c.client == nil{
		return errors.New("redis client can not be empty")
	}
	if c.topic == "" || c.consumerID == "" || c.groupID == "" {
		return errors.New("topic | group_id | consumer_id can't be empty")
	}

	return nil
}

func NewConsumer(client *redis.Client, topic , groupID,consumerID string, callbackFunc MsgCallback, opts ...ConsumerOption)(*Consumer, error){
	ctx,stop := context.WithCancel(context.Background())
	c := Consumer{
		client: client,
		ctx: ctx,
		stop: stop,
		callbackFunc: callbackFunc,
		topic: topic,
		groupID: groupID,
		consumerID: consumerID,
		opts: &ConsumerOptions{},
		failureCnts: make(map[redis.MsgEntity]int),
	}
	if err := c.checkParam();err != nil{
		return nil,err
	}

	for _,opt := range opts{
		opt(c.opts)
	}
	repairConsumer(c.opts)
	go c.run()
	return &c,nil
}

func (c *Consumer)Stop(){
	c.stop()
}

func (c *Consumer)run(){
	for {
		select {
		case <- c.ctx.Done():
			return
		default:

		}

		msgs,err := c.receive()
		if err != nil{
			log.ErrorContextf(c.ctx, "receive msg failed, err: %v", err)
			continue
		}

		tctx,_ := context.WithTimeout(c.ctx, c.opts.handleMsgsTimeout)
		c.handlerMsgs(tctx, msgs)

		tctx, _ = context.WithTimeout(c.ctx, c.opts.deadLetterDeliverTimeout)
		c.deliverDeadLetter(tctx)

		// pending 消息接收处理
		pendingMsgs, err := c.receivePending()
		if err != nil {
			log.ErrorContextf(c.ctx, "pending msg received failed, err: %v", err)
			continue
		}

		tctx, _ = context.WithTimeout(c.ctx, c.opts.handleMsgsTimeout)
		c.handlerMsgs(tctx, pendingMsgs)
	}
}

func (c *Consumer)receive()([]*redis.MsgEntity, error){
	msgs,err := c.client.XReadGroup(c.ctx, c.groupID, c.consumerID, c.topic,int(c.opts.receiveTimeout.Milliseconds()))
	if err != nil && !errors.Is(err, redis.ErrNoMsg){
		return nil,err
	}
	return msgs,nil
}

func (c *Consumer)receivePending()([]*redis.MsgEntity, error){
	msgs,err := c.client.XReadGroupPending(c.ctx, c.groupID, c.consumerID, c.topic)
	if err != nil && !errors.Is(err, redis.ErrNoMsg){
		return nil,err
	}
	return msgs,nil
}

func (c *Consumer)handlerMsgs(ctx context.Context, msgs []*redis.MsgEntity){
	for _,msg := range  msgs{
		if err := c.callbackFunc(ctx, msg); err != nil{
			c.failureCnts[*msg] ++
			continue
		}

		if err := c.client.XACK(ctx, c.topic, c.groupID, msg.MsgID);err != nil{
			log.ErrorContextf(ctx, "msg ack failed, msg id: %s, err: %v", msg.MsgID, err)
			continue
		}
		delete(c.failureCnts, *msg)
	}
}
func (c *Consumer) deliverDeadLetter(ctx context.Context) {
	// 对于失败达到指定次数的消息，投递到死信中，然后执行 ack
	for msg, failureCnt := range c.failureCnts {
		if failureCnt < c.opts.maxRetryLimit {
			continue
		}

		// 投递死信队列
		if err := c.opts.deadLetterMailbox.Deliver(ctx, &msg); err != nil {
			log.ErrorContextf(c.ctx, "dead letter deliver failed, msg id: %s, err: %v", msg.MsgID, err)
		}

		// 执行 ack 响应
		if err := c.client.XACK(ctx, c.topic, c.groupID, msg.MsgID); err != nil {
			log.ErrorContextf(c.ctx, "msg ack failed, msg id: %s, err: %v", msg.MsgID, err)
			continue
		}

		// 对于 ack 成功的消息，将其从 failure map 中删除
		delete(c.failureCnts, msg)
	}
}
