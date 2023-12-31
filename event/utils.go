package event

import (
	"time"
)

func TestTopicConfigs(topicName string, brokers ...string) *TopicConfigs {
	return &TopicConfigs{
		Topic:                 topicName,
		Brokers:               brokers,
		NumPartitions:         1,
		Partition:             0,
		ReplicationFactor:     1,
		WaitMaxTime:           2 * time.Second,
		AwaitBetweenReadsTime: 500 * time.Millisecond,
		AutoCreateTopic:       true,
	}
}
