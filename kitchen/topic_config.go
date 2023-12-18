package kitchen

import (
	"github.com/spf13/cast"
	"mc-burger-orders/event"
	"mc-burger-orders/log"
	"os"
)

func TopicConfigsFromEnv() *event.TopicConfigs {
	topic := os.Getenv("KAFKA_TOPICS__KITCHEN_REQUESTS_TOPIC_NAME")
	if len(topic) <= 0 {
		log.Error.Panicf("Kafka Topic `kitchen request events handler` name is missing")
	}

	partition := 3
	partitionVal := os.Getenv("KAFKA_TOPICS__KITCHEN_REQUESTS_PARTITION")
	numPartitionsVal := os.Getenv("KAFKA_TOPICS__KITCHEN_REQUESTS_NUMBER_OF_PARTITIONS")
	replicationFactorVal := os.Getenv("KAFKA_TOPICS__KITCHEN_REQUESTS_REPLICA_FACTOR")

	if len(partitionVal) > 0 {
		partition = cast.ToInt(partitionVal)
	}
	return event.NewTopicConfig(topic, partition, numPartitionsVal, replicationFactorVal)
}
