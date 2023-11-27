package order

import (
	"github.com/spf13/cast"
	"mc-burger-orders/event"
	"os"
	"strings"
	"time"
)

func StatusUpdatedTopicConfigsFromEnv() *event.TopicConfigs {
	kafkaAddressEnvVal := os.Getenv("KAFKA_ADDRESS")
	kafkaAddress := strings.Split(kafkaAddressEnvVal, ",")
	topic := os.Getenv("KAFKA_TOPICS__ORDER_STATUS_TOPIC_NAME")

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
		ReplicationFactor:     replicationFactor,
		WaitMaxTime:           10 * time.Second,
		AwaitBetweenReadsTime: 2 * time.Second,
	}
}
