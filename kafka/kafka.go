package kafka

import (
	"github.com/Shopify/sarama"
)

func NewClusterConsumer(brokers []string, group string) (cli sarama.ConsumerGroup, err error) {
	cfg := sarama.NewConfig()

	cfg.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	cfg.Consumer.Offsets.Initial = sarama.OffsetNewest
	cfg.ChannelBufferSize = 1000

	cli, err = sarama.NewConsumerGroup(brokers, group, cfg)
	return
}
