package kafka

import (
	"github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"
)

func NewClusterConsumer(addrs []string, group string, topics []string) *cluster.Consumer {
	cfg := cluster.NewConfig()

	cfg.Consumer.Offsets.Initial = sarama.OffsetNewest

	consumer, err := cluster.NewConsumer(addrs, group, topics, cfg)
	if err != nil {
		panic(err.Error())
	}
	return consumer
}
