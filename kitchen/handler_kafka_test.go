package kitchen

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"mc-burger-orders/event"
	"mc-burger-orders/shelf"
	"mc-burger-orders/testing/data"
	"mc-burger-orders/testing/utils"
	"strconv"
	"testing"
	"time"
)

var (
	kafkaConfig *event.TopicConfigs
	testStack   *shelf.Shelf
	eventBus    event.EventBus
	topic       = fmt.Sprintf("test-kitchen-requests-%d", rand.Intn(100))
)

func TestIntegrationHandler_WithKafkaMessages(t *testing.T) {
	utils.IntegrationTest(t)
	ctx := context.Background()
	testStack = shelf.NewEmptyShelf()
	kafkaContainer, brokers := utils.TestWithKafka(t, ctx)
	kafkaConfig = event.TestTopicConfigs(topic, brokers...)
	eventBus = event.NewInternalEventBus()

	t.Run("should submit new request to worker when message arrives", shouldSubmitItemRequestToWorkerWhenMessageArrives)

	t.Cleanup(func() {
		t.Log("Running Clean UP code")
		utils.TerminateKafka(t, ctx, kafkaContainer)
	})
}

func shouldSubmitItemRequestToWorkerWhenMessageArrives(t *testing.T) {
	// given
	msgChan := make(chan kafka.Message)
	msg := make([]map[string]any, 0)
	msg = data.AppendHamburgerItem(msg, 1)

	msg2 := make([]map[string]any, 0)
	msg2 = data.AppendSpicyStripesItem(msg2, 2)

	for range make([]int, 10) {
		go sendMessages(t, msg)
		go sendMessages(t, msg2)
	}

	commandHandler := NewHandler(testStack)
	eventBus.AddHandler(commandHandler)

	// when
	reader := event.NewTopicReader(kafkaConfig, eventBus)
	go reader.SubscribeToTopic(msgChan)

	// then
	cnt := 0
	for {
		time.Sleep(1 * time.Second)
		if assertExpectedItemsCreated(t) {
			return
		}
		t.Log("not all items created yet!")
		cnt++
		if cnt > 25 {
			assert.Fail(t, "not all items created")
			return
		}
	}
}

func assertExpectedItemsCreated(t *testing.T) bool {
	hamburgers := testStack.GetCurrent("hamburger")
	spicyStripes := testStack.GetCurrent("spicy-stripes")
	t.Logf("Hamburgers: %d || Spicy-stripes: %d", hamburgers, spicyStripes)
	return 10 <= hamburgers && 20 <= spicyStripes
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
	err := writer.SendMessage(context.Background(), msg)
	if err != nil {
		assert.Fail(t, "failed to send message to topic", kafkaConfig.Topic)
	}
}
