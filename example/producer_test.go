package example

import (
	"context"
	"learn_redis_mq"
	"learn_redis_mq/redis"
	"testing"


)

func Test_Producer(t *testing.T) {
	client := redis.NewClient(network, address, password)
	// 最多保留十条消息
	producer := learn_redis_mq.NewProducer(client, learn_redis_mq.WithMsgQueueLen(10))
	ctx := context.Background()
	msgID, err := producer.SendMsg(ctx, topic, "test_kk", "test_vv")
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(msgID)
}
