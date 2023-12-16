package order

import (
	"github.com/spf13/cast"
	"mc-burger-orders/event"
	"mc-burger-orders/log"
	"os"
)

func StatusUpdatedEndpointTopicConfigsFromEnv() *event.TopicConfigs {
	partition := 3
	partitionsVal := os.Getenv("KAFKA_TOPICS__ORDER_STATUS_ENDPOINT_PARTITION")
	if len(partitionsVal) > 0 {
		partition = cast.ToInt(partitionsVal)
	}
	return defaultStatusTopicConfig(partition)
}

func StatusUpdatedTopicConfigsFromEnv() *event.TopicConfigs {
	partition := 0
	partitionsVal := os.Getenv("KAFKA_TOPICS__ORDER_STATUS_PARTITION")
	if len(partitionsVal) > 0 {
		partition = cast.ToInt(partitionsVal)
	}
	return defaultStatusTopicConfig(partition)
}

func defaultStatusTopicConfig(partition int) *event.TopicConfigs {
	topic := os.Getenv("KAFKA_TOPICS__ORDER_STATUS_TOPIC_NAME")
	if len(topic) <= 0 {
		log.Error.Panicf("Kafka Topic `order status events handler` name is missing")
	}

	numPartitionsVal := os.Getenv("KAFKA_TOPICS__ORDER_STATUS_NUMBER_OF_PARTITIONS")
	replicationFactorVal := os.Getenv("KAFKA_TOPICS__ORDER_STATUS_REPLICA_FACTOR")
	return event.NewTopicConfig(topic, partition, numPartitionsVal, replicationFactorVal)
}
