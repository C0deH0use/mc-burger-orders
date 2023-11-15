package kitchen

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"log"
	"math/rand"
	"mc-burger-orders/event"
	stack2 "mc-burger-orders/stack"
	"mc-burger-orders/testing/utils"
	"strconv"
	"testing"
	"time"
)

var (
	kafkaConfig *event.TopicConfigs
	testStack   *stack2.Stack
	eventBus    event.EventBus
	topic       = fmt.Sprintf("test-kitchen-requests-%d", rand.Intn(100))
)

func TestCommandsHandler_WithKafkaMessages(t *testing.T) {

	ctx := context.Background()
	testStack = stack2.NewEmptyStack()
	kafkaContainer, brokers := utils.TestWithKafka(ctx)
	kafkaConfig = &event.TopicConfigs{
		Topic:             topic,
		Brokers:           brokers,
		NumPartitions:     1,
		ReplicationFactor: 1,
		AutoCreateTopic:   true,
	}

	t.Run("should submit new request to worker when message arrives", shouldSubmitItemRequestToWorkerWhenMessageArrives)

	t.Cleanup(func() {
		log.Println("Running Clean UP code")
		utils.TerminateKafka(kafkaContainer)
	})
}

func shouldSubmitItemRequestToWorkerWhenMessageArrives(t *testing.T) {
	// given
	msgChan := make(chan kafka.Message)
	msg := make([]map[string]any, 0)
	msg = append(msg, map[string]any{
		"itemName": "hamburger",
		"quantity": 2,
	})

	for range make([]int, 10) {
		go sendMessages(t, msg)
	}

	commandHandler := NewHandler(kafkaConfig, kafkaConfig, testStack)
	go commandHandler.AwaitOn(msgChan)

	// when
	reader := event.NewTopicReader(kafkaConfig, eventBus)
	reader.SubscribeToTopic(msgChan)

	// then
	for {
		currentHamburgers := testStack.GetCurrent("hamburger")

		if currentHamburgers >= 10 {
			return
		}
	}
}

func sendMessages(t *testing.T, requestPayload []map[string]any) {
	writer := event.NewTopicWriter(kafkaConfig)

	msgKey := []byte(strconv.FormatInt(time.Now().UnixNano(), 10))

	headers := make([]kafka.Header, 0)
	headers = append(headers, kafka.Header{Key: "order", Value: []byte("1010")})
	headers = append(headers, kafka.Header{Key: "event", Value: []byte("request-item")})

	b, _ := json.Marshal(requestPayload)

	msg := kafka.Message{
		Key:     msgKey,
		Headers: headers,
		Value:   b,
	}

	// when
	log.Printf("Sending message on topic %v", kafkaConfig.Topic)

	err := writer.SendMessage(context.Background(), msg)
	if err != nil {
		assert.Fail(t, "failed to send message to topic", kafkaConfig.Topic)
	}
}
