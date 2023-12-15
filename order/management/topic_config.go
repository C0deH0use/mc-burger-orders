package management

import (
	"github.com/spf13/cast"
	"mc-burger-orders/event"
	"mc-burger-orders/log"
	"os"
)

func OrderJobsTopicConfigsFromEnv() *event.TopicConfigs {
	topic := os.Getenv("KAFKA_TOPICS__ORDER_JOBS_TOPIC_NAME")
	if len(topic) <= 0 {
		log.Error.Panicf("Kafka Topic `order status events handler` name is missing")
	}
	partition := 0
	partitionsVal := os.Getenv("KAFKA_TOPICS__ORDER_JOBS_CHECK_PARTITION")
	if len(partitionsVal) > 0 {
		partition = cast.ToInt(partitionsVal)
	}

	numPartitionsVal := os.Getenv("KAFKA_TOPICS__ORDER_JOBS_CHECK_NUMBER_OF_PARTITIONS")
	replicationFactorVal := os.Getenv("KAFKA_TOPICS__ORDER_JOBS_CHECK_REPLICA_FACTOR")
	return event.NewTopicConfig(topic, partition, numPartitionsVal, replicationFactorVal)
}
