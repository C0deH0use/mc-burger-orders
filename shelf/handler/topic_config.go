package handler

import (
	"github.com/spf13/cast"
	"mc-burger-orders/event"
	"mc-burger-orders/log"
	"os"
	"strings"
	"time"
)

func TopicConfigsFromEnv() *event.TopicConfigs {
	kafkaAddressEnvVal := os.Getenv("KAFKA_ADDRESS")
	kafkaAddress := strings.Split(kafkaAddressEnvVal, ",")
	topic := os.Getenv("KAFKA_TOPICS__SHELF_HANDLER_TOPIC_NAME")
	if len(topic) <= 0 {
		log.Error.Panicf("Kafka Topic `Shelf Handler` name is missing")
	}

	numPartitions := 3
	numPartitionsVal := os.Getenv("KAFKA_TOPICS__SHELF_HANDLER_TOPIC_NUMBER_OF_PARTITIONS")
	replicationFactor := 1
	replicationFactorVal := os.Getenv("KAFKA_TOPICS__SHELF_HANDLER_TOPIC_REPLICA_FACTOR")

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