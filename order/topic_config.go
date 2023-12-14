package order

import (
	"github.com/spf13/cast"
	"mc-burger-orders/event"
	"mc-burger-orders/log"
	"os"
	"strings"
	"time"
)

func StatusUpdatedEndpointTopicConfigsFromEnv() *event.TopicConfigs {
	partition := 3
	partitionsVal := os.Getenv("KAFKA_TOPICS__ORDER_STATUS_ENDPOINT_PARTITION")
	if len(partitionsVal) > 0 {
		partition = cast.ToInt(partitionsVal)
	}
	return defaultTopicConfig(partition)
}

func StatusUpdatedTopicConfigsFromEnv() *event.TopicConfigs {
	partition := 0
	partitionsVal := os.Getenv("KAFKA_TOPICS__ORDER_STATUS_PARTITION")
	if len(partitionsVal) > 0 {
		partition = cast.ToInt(partitionsVal)
	}
	return defaultTopicConfig(partition)
}

func defaultTopicConfig(partition int) *event.TopicConfigs {
	kafkaAddressEnvVal := os.Getenv("KAFKA_ADDRESS")
	kafkaAddress := strings.Split(kafkaAddressEnvVal, ",")
	topic := os.Getenv("KAFKA_TOPICS__ORDER_STATUS_TOPIC_NAME")
	if len(topic) <= 0 {
		log.Error.Panicf("Kafka Topic `order status events handler` name is missing")
	}

	numPartitions := 3
	numPartitionsVal := os.Getenv("KAFKA_TOPICS__ORDER_STATUS_NUMBER_OF_PARTITIONS")
	replicationFactor := 1
	replicationFactorVal := os.Getenv("KAFKA_TOPICS__ORDER_STATUS_REPLICA_FACTOR")

	if len(numPartitionsVal) > 0 {
		numPartitions = cast.ToInt(numPartitionsVal)
	}
	if len(replicationFactorVal) > 0 {
		replicationFactor = cast.ToInt(replicationFactorVal)
	}

	return &event.TopicConfigs{
		Brokers:               kafkaAddress,
		Topic:                 topic,
		NumPartitions:         numPartitions,
		Partition:             partition,
		ReplicationFactor:     replicationFactor,
		WaitMaxTime:           2 * time.Second,
		AwaitBetweenReadsTime: 500 * time.Millisecond,
	}
}
