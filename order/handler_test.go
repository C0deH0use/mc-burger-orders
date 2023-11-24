package order

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"math/rand"
	"mc-burger-orders/event"
	"mc-burger-orders/kitchen/item"
	m "mc-burger-orders/order/model"
	"mc-burger-orders/stack"
	"mc-burger-orders/testing/utils"
	"strconv"
	"testing"
	"time"
)

var (
	stackKafkaConfig       *event.TopicConfigs
	orderStatusKafkaConfig *event.TopicConfigs
	stackTopic             = fmt.Sprintf("test-stack-events-%d", rand.Intn(100))
	orderStatusTopic       = fmt.Sprintf("test-order-status-events-%d", rand.Intn(100))
)

func TestOrdersHandler_Handle(t *testing.T) {
	ctx := context.Background()

	mongoContainer = utils.TestWithMongo(t, ctx)
	kafkaContainer, brokers := utils.TestWithKafka(t, ctx)
	kitchenRequestsKafkaConfig = event.TestTopicConfigs(topic, brokers...)
	stackKafkaConfig = event.TestTopicConfigs(stackTopic, brokers...)
	orderStatusKafkaConfig = event.TestTopicConfigs(orderStatusTopic, brokers...)

	database = utils.GetMongoDbFrom(t, mongoContainer)
	collectionDb = database.Collection("orders")
	orderNumberCollectionDb = database.Collection("order-numbers")

	t.Run("should pack item when stack event occurred", shouldPackPreparedItemWhenEvenFromStackOccurred)

	t.Cleanup(func() {
		log.Println("Running Clean UP code")
		utils.TerminateMongo(t, mongoContainer)
		utils.TerminateKafka(t, kafkaContainer)
	})
}

func shouldPackPreparedItemWhenEvenFromStackOccurred(t *testing.T) {
	// given
	kitchenStack := stack.NewEmptyStack()
	expectedOrders := []interface{}{
		m.Order{OrderNumber: 999, CustomerId: 10, Items: []item.Item{{Name: "hamburger", Quantity: 1}, {Name: "fries", Quantity: 1}}, Status: m.Ready, CreatedAt: time.Now(), ModifiedAt: time.Now()},
		m.Order{OrderNumber: 1000, CustomerId: 1, Items: []item.Item{{Name: "hamburger", Quantity: 1}, {Name: "fries", Quantity: 1}}, Status: m.Requested, CreatedAt: time.Now(), ModifiedAt: time.Now()},
		m.Order{OrderNumber: 1002, CustomerId: 3, Items: []item.Item{{Name: "cheeseburger", Quantity: 2}, {Name: "hamburger", Quantity: 3}}, Status: m.Requested, CreatedAt: time.Now(), ModifiedAt: time.Now()},
	}
	utils.DeleteMany(t, collectionDb, bson.D{})
	utils.InsertMany(t, collectionDb, expectedOrders)

	t.Log("Creating EventBUS")
	eventBus := event.NewInternalEventBus()
	ordersHandler := NewHandler(database, kitchenRequestsKafkaConfig, orderStatusKafkaConfig, kitchenStack)

	t.Log("Configuring EventBUS with order handler")
	eventBus.AddHandler(ordersHandler)
	kitchenStack.Add("fries")
	kitchenStack.Add("hamburger")

	payload := make([]map[string]any, 0)
	payload = append(payload, map[string]any{
		"itemName": "hamburger",
		"quantity": 1,
	})
	payload = append(payload, map[string]any{
		"itemName": "fries",
		"quantity": 1,
	})

	t.Log("Configuring EventBUS with order handler")
	sendItemAddedToStackMessages(t, payload)

	// when
	stackTopicReader := event.NewTopicReader(stackKafkaConfig, eventBus)
	go stackTopicReader.SubscribeToTopic(make(chan kafka.Message))

	// then
	checkCnt := 0
	for {
		order := fetchByOrderNumber(t, 1000)
		if order.Status == m.Ready {
			return
		}

		if checkCnt > 6 {
			assert.Fail(t, "Order Status not changed to Ready")
			return
		}

		time.Sleep(2 * time.Second)
		checkCnt++
	}
}

func sendItemAddedToStackMessages(t *testing.T, requestPayload []map[string]any) {
	t.Log("Configure TestWriter for topic", stackKafkaConfig.Topic)
	writer := event.NewTopicWriter(stackKafkaConfig)

	msgKey := []byte(strconv.FormatInt(time.Now().UnixNano(), 10))

	headers := make([]kafka.Header, 0)
	headers = append(headers, kafka.Header{Key: "event", Value: []byte(stack.ItemAddedToStackEvent)})

	b, _ := json.Marshal(requestPayload)

	msg := kafka.Message{
		Key:     msgKey,
		Headers: headers,
		Value:   b,
	}

	// when
	t.Log("Sending TestMessage to topic")
	err := writer.SendMessage(context.Background(), msg)
	if err != nil {
		assert.Fail(t, "failed to send message to topic", stackKafkaConfig.Topic)
	}
}

func fetchByOrderNumber(t *testing.T, orderNumber int64) *m.Order {
	filter := bson.D{{Key: "orderNumber", Value: orderNumber}}
	result := collectionDb.FindOne(context.Background(), filter)
	if result.Err() != nil {
		assert.Fail(t, "failed to fetch record")
	}

	var order *m.Order
	if err := result.Decode(&order); err != nil {
		assert.Fail(t, "failed to decoded record")
	}

	return order
}
