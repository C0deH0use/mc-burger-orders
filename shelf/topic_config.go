package shelf

import (
	"github.com/spf13/cast"
	"mc-burger-orders/event"
	"os"
	"strings"
	"time"
)

func TopicConfigsFromEnv() *event.TopicConfigs {
	kafkaAddressEnvVal := os.Getenv("KAFKA_ADDRESS")
	kafkaAddress := strings.Split(kafkaAddressEnvVal, ",")
	topic := os.Getenv("KAFKA_TOPICS__STACK_TOPIC_NAME")

	numPartitions := 3
	numPartitionsVal := os.Getenv("KAFKA_TOPICS__STACK_TOPIC_NUMBER_OF_PARTITIONS")
	replicationFactor := 1
	replicationFactorVal := os.Getenv("KAFKA_TOPICS__STACK_TOPIC_REPLICA_FACTOR")

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
