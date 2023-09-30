package learn_redis_mq

import (
	"context"
	"learn_redis_mq/log"
	"learn_redis_mq/redis"
)

type DeadLetterMailbox interface {
	Deliver(ctx context.Context, msg *redis.MsgEntity)error
}

type DeadLetterLogger struct {

}

func NewDeadLetterLogger()*DeadLetterLogger{
	return &DeadLetterLogger{}
}

func(d *DeadLetterLogger)Deliver(ctx context.Context, msg *redis.MsgEntity)error{
	log.ErrorContextf(ctx, "msg fail execeed retry limit, msg id: %s", msg.MsgID)
	return nil
}