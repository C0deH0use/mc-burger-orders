package shelf

import (
	"github.com/spf13/cast"
	"mc-burger-orders/event"
	"mc-burger-orders/log"
	"os"
)

func TopicConfigsFromEnv() *event.TopicConfigs {
	topic := os.Getenv("KAFKA_TOPICS__SHELF_TOPIC_NAME")
	if len(topic) <= 0 {
		log.Error.Panicf("Kafka Topic `Shelf events handler` name is missing")
	}

	partition := 0
	partitionVal := os.Getenv("KAFKA_TOPICS__SHELF_TOPIC_PARTITION")
	if len(partitionVal) > 0 {
		partition = cast.ToInt(partitionVal)
	}

	numPartitionsVal := os.Getenv("KAFKA_TOPICS__SHELF_TOPIC_NUMBER_OF_PARTITIONS")
	replicationFactorVal := os.Getenv("KAFKA_TOPICS__SHELF_TOPIC_REPLICA_FACTOR")

	return event.NewTopicConfig(topic, partition, numPartitionsVal, replicationFactorVal)
}
