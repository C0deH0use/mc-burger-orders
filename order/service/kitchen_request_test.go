package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"mc-burger-orders/event"
	"mc-burger-orders/utils"
	"strconv"
	"testing"
	"time"
)

var (
	sut        *KitchenService
	ctx        context.Context
	testReader *kafka.Reader
	topic      = fmt.Sprintf("test-kitchen-requests-%d", rand.Intn(100))
)

func TestKitchenService_RequestForOrder(t *testing.T) {
	ctx = context.Background()
	kafkaContainer := utils.TestWithKafka(ctx)
	brokers, err := kafkaContainer.Brokers(ctx)
	if err != nil {
		assert.Fail(t, "cannot read Brokers from kafka container")
	}
	kafkaConfig := &event.TopicConfigs{
		Topic:             topic,
		Brokers:           brokers,
		NumPartitions:     1,
		ReplicationFactor: 1,
		AutoCreateTopic:   true,
	}
	sut = NewKitchenServiceFrom(kafkaConfig)
	testReader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:   brokers,
		Topic:     topic,
		Partition: 0,
		MinBytes:  10e3, // 10KB
		MaxBytes:  10e6, // 10MB
		MaxWait:   time.Millisecond * 10,
	})

	t.Run("Run KitchenService Tests", func(t *testing.T) {
		t.Run("should correctly send new request to the kitchen", shouldSendNewMessageToTopic)
	})

	// Clean up the container after
	t.Cleanup(func() {
		utils.TerminateKafka(kafkaContainer)
	})
}

func shouldSendNewMessageToTopic(t *testing.T) {
	// given
	itemName := "hamburger"
	quantity := 2
	orderNumber := int64(122)

	// when
	err := sut.RequestForOrder(ctx, itemName, quantity, orderNumber)

	// then
	assert.Nil(t, err)

	// and
	expectedHeader := kafka.Header{Key: "order", Value: []byte(strconv.FormatInt(orderNumber, 10))}
	expectedMessage := NewKitchenRequestMessage(itemName, quantity)
	message, err := testReader.ReadMessage(context.Background())

	if err != nil {
		assert.Fail(t, "failed reading message on test Topic", err)
	}

	// and
	assert.Equal(t, topic, message.Topic)
	assert.Len(t, message.Headers, 1)
	assert.Equal(t, expectedHeader, message.Headers[0])

	actualMessage := &KitchenRequestMessage{}
	err = json.Unmarshal(message.Value, actualMessage)
	if err != nil {
		assert.Fail(t, "failed to unmarshal message on test Topic", err)
	}
	assert.Equal(t, expectedMessage, actualMessage)
}
