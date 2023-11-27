package event

import "time"

func TestTopicConfigs(topicName string, brokers ...string) *TopicConfigs {
	return &TopicConfigs{
		Topic:                 topicName,
		Brokers:               brokers,
		NumPartitions:         1,
		ReplicationFactor:     1,
		WaitMaxTime:           500 * time.Millisecond,
		AwaitBetweenReadsTime: 500 * time.Millisecond,
		AutoCreateTopic:       true,
	}
}
