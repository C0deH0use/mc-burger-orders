package order

import (
	"github.com/spf13/cast"
	"mc-burger-orders/event"
	"mc-burger-orders/log"
	"os"
)

func StatusUpdatedEndpointTopicConfigsFromEnv() *event.TopicConfigs {
	partition := 1
	partitionsVal := os.Getenv("KAFKA_TOPICS__ORDER_STATUS_ENDPOINT_PARTITION")
	if len(partitionsVal) > 0 {
		partition = cast.ToInt(partitionsVal)
	}
	return defaultStatusTopicConfig(partition)
}

func StatusUpdatedTopicConfigsFromEnv() *event.TopicConfigs {
	partition := 0
	partitionsVal := os.Getenv("KAFKA_TOPICS__ORDER_STATUS_SERVICE_PARTITION")
	if len(partitionsVal) > 0 {
		partition = cast.ToInt(partitionsVal)
	}
	return defaultStatusTopicConfig(partition)
}

func StreamTopicConfigsFromEnv() *event.TopicConfigs {
	topic := os.Getenv("KAFKA_TOPICS__ORDER_STREAM_TOPIC_NAME")
	if len(topic) <= 0 {
		log.Error.Panicf("Kafka Topic `order stream` name is missing")
	}

	partition := 1
	partitionsVal := os.Getenv("KAFKA_TOPICS__ORDER_STREAM_PARTITION")
	if len(partitionsVal) > 0 {
		partition = cast.ToInt(partitionsVal)
	}
	numPartitionsVal := os.Getenv("KAFKA_TOPICS__ORDER_STREAM_NUMBER_OF_PARTITIONS")
	replicationFactorVal := os.Getenv("KAFKA_TOPICS__ORDER_STREAM_REPLICA_FACTOR")
	return event.NewTopicConfig(topic, partition, numPartitionsVal, replicationFactorVal)
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
