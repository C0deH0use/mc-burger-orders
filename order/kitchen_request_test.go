package order

import (
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"mc-burger-orders/event"
	"mc-burger-orders/order/dto"
	"mc-burger-orders/testing/utils"
	"strconv"
	"testing"
	"time"
)

var (
	sut        *KitchenService
	ctx        context.Context
	testReader *kafka.Reader
)

func TestIntegrationKitchenService_RequestForOrder(t *testing.T) {
	utils.IntegrationTest(t)
	ctx = context.Background()
	kafkaContainer, brokers := utils.TestWithKafka(t, ctx)
	kafkaConfig := event.TestTopicConfigs(topic, brokers...)
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
		utils.TerminateKafka(t, ctx, kafkaContainer)
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
	expectedOrderHeader := kafka.Header{Key: "order", Value: []byte(strconv.FormatInt(orderNumber, 10))}
	expectedEventHeader := kafka.Header{Key: "event", Value: []byte("request-item")}
	expectedMessage := make([]*dto.KitchenRequestMessage, 0)
	expectedMessage = append(expectedMessage, dto.NewKitchenRequestMessage(itemName, quantity))
	message, err := testReader.ReadMessage(context.Background())

	if err != nil {
		assert.Fail(t, "failed reading message on test Topic", err)
	}

	// and
	assert.Equal(t, topic, message.Topic)
	assert.Len(t, message.Headers, 2)
	assert.Contains(t, message.Headers, expectedOrderHeader)
	assert.Contains(t, message.Headers, expectedEventHeader)

	actualMessage := make([]*dto.KitchenRequestMessage, 0)
	actualMessage = append(actualMessage, dto.NewKitchenRequestMessage(itemName, quantity))

	err = json.Unmarshal(message.Value, &actualMessage)
	if err != nil {
		assert.Fail(t, "failed to unmarshal message on test Topic", err)
	}
	assert.Equal(t, expectedMessage, actualMessage)
}
