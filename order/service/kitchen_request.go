package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/segmentio/kafka-go"
	"github.com/spf13/cast"
	"mc-burger-orders/event"
	"mc-burger-orders/log"
	"os"
	"strconv"
	"strings"
	"time"
)

type KitchenRequestService interface {
	RequestForOrder(ctx context.Context, itemName string, quantity int, orderNumber int64) error
}

type KitchenService struct {
	*event.DefaultWriter
}

func KitchenTopicConfigsFromEnv() *event.TopicConfigs {
	kafkaAddressEnvVal := os.Getenv("KAFKA_ADDRESS")
	kafkaAddress := strings.Split(kafkaAddressEnvVal, ",")
	topic := os.Getenv("KAFKA_TOPICS__KITCHEN_REQUESTS_TOPIC_NAME")

	numPartitions := 3
	numPartitionsVal := os.Getenv("KAFKA_TOPICS__KITCHEN_REQUESTS_TOPIC_NUMBER_OF_PARTITIONS")
	replicationFactor := 1
	replicationFactorVal := os.Getenv("KAFKA_TOPICS__KITCHEN_REQUESTS_TOPIC_REPLICA_FACTOR")

	if len(numPartitionsVal) > 0 {
		numPartitions = cast.ToInt(numPartitionsVal)
	}
	if len(replicationFactorVal) > 0 {
		replicationFactor = cast.ToInt(replicationFactorVal)
	}

	return &event.TopicConfigs{
		Brokers:           kafkaAddress,
		Topic:             topic,
		AutoCreateTopic:   true,
		NumPartitions:     numPartitions,
		ReplicationFactor: replicationFactor,
	}
}

func NewKitchenServiceFrom(config *event.TopicConfigs) *KitchenService {
	defaultWriter := event.NewTopicWriter(config)
	return &KitchenService{defaultWriter}
}

func (s *KitchenService) RequestForOrder(ctx context.Context, itemName string, quantity int, orderNumber int64) error {
	headers := make([]kafka.Header, 0)
	headers = append(headers, kafka.Header{Key: "order", Value: []byte(strconv.FormatInt(orderNumber, 10))})
	message := NewKitchenRequestMessage(itemName, quantity)
	msgValue, err := json.Marshal(message)
	if err != nil {
		err = fmt.Errorf("failed to convert message details to bytes. Reason: %s", err)
		return err
	}

	msgKey := []byte(strconv.FormatInt(time.Now().UnixNano(), 10))
	log.Info.Println("Sending message with value", string(msgValue))
	msg := kafka.Message{
		Headers: headers,
		Key:     msgKey,
		Value:   msgValue,
	}
	err = s.SendMessage(ctx, msg)
	if err != nil {
		return err
	}
	return nil
}
