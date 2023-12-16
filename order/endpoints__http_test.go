package order

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/segmentio/kafka-go"
	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"math/rand"
	"mc-burger-orders/event"
	"mc-burger-orders/kitchen/item"
	s "mc-burger-orders/order/dto"
	"mc-burger-orders/shelf"
	"mc-burger-orders/testing/utils"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

var (
	mongoContainer             *mongodb.MongoDBContainer
	database                   *mongo.Database
	collectionDb               *mongo.Collection
	orderNumberCollectionDb    *mongo.Collection
	kitchenRequestsKafkaConfig *event.TopicConfigs
	kitchenRequestsReader      *kafka.Reader
	topic                      = fmt.Sprintf("test-kitchen-requests-%d", rand.Intn(100))
)

func TestIntegrationOrder_HttpEndpoints(t *testing.T) {
	utils.IntegrationTest(t)
	ctx := context.Background()
	mongoContainer, database = utils.TestWithMongo(t, ctx)
	kafkaContainer, brokers := utils.TestWithKafka(t, ctx)
	kitchenRequestsKafkaConfig = event.TestTopicConfigs(topic, brokers...)
	orderStatusKafkaConfig = event.TestTopicConfigs(orderStatusTopic, brokers...)

	collectionDb = database.Collection("orders")
	orderNumberCollectionDb = database.Collection("order-numbers")

	time.Sleep(time.Second * 5)

	t.Run("should return orders", shouldFetchOrdersWhenMultipleStored)
	t.Run("should store and begin packing order when received valid request", shouldBeginPackingAndStoreOrderWhenRequested)

	t.Cleanup(func() {
		t.Log("Running Clean UP code")
		utils.TerminateMongo(t, ctx, mongoContainer)
		utils.TerminateKafka(t, ctx, kafkaContainer)
	})
}

func shouldFetchOrdersWhenMultipleStored(t *testing.T) {
	// given
	expectedOrders := []interface{}{
		Order{OrderNumber: 1000, CustomerId: 1, Items: []item.Item{{Name: "hamburger", Quantity: 1}, {Name: "fries", Quantity: 1}}, Status: Ready, CreatedAt: time.Now(), ModifiedAt: time.Now()},
		Order{OrderNumber: 1001, CustomerId: 2, Items: []item.Item{{Name: "hamburger", Quantity: 1}, {Name: "cheeseburger", Quantity: 2}}, Status: InProgress, CreatedAt: time.Now(), ModifiedAt: time.Now()},
		Order{OrderNumber: 1002, CustomerId: 3, Items: []item.Item{{Name: "cheeseburger", Quantity: 2}, {Name: "cheeseburger", Quantity: 3}}, Status: Requested, CreatedAt: time.Now(), ModifiedAt: time.Now()},
	}
	utils.DeleteMany(t, collectionDb, bson.D{})
	utils.InsertMany(t, collectionDb, expectedOrders)

	endpoints := NewOrderEndpoints(database, kitchenRequestsKafkaConfig, orderStatusKafkaConfig, shelf.NewEmptyShelf())
	engine := utils.SetUpRouter(endpoints.Setup)

	req, _ := http.NewRequest("GET", "/order", nil)
	resp := httptest.NewRecorder()

	// when
	engine.ServeHTTP(resp, req)

	// then
	assert.Equal(t, http.StatusOK, resp.Code)

	// and
	var payload []map[string]any
	err := json.Unmarshal(resp.Body.Bytes(), &payload)
	if err != nil {
		assert.Fail(t, "Error while unmarshalling response payload to map", err)
	}

	assert.Equal(t, 3, len(payload))

	orderOne := payload[0]
	assert.Equal(t, 1000.0, orderOne["orderNumber"])
	assert.Equal(t, 1.0, orderOne["customerId"])
	assert.Equal(t, "READY", orderOne["status"])

	defer func() {
		utils.DeleteMany(t, collectionDb, bson.D{})
		utils.DeleteMany(t, orderNumberCollectionDb, bson.D{})
	}()
}

func shouldBeginPackingAndStoreOrderWhenRequested(t *testing.T) {
	// given
	order := &map[string]any{
		"customerId": 10,
		"items": []interface{}{
			map[string]any{
				"name":     "hamburger",
				"quantity": 2,
			},
			map[string]any{
				"name":     "cheeseburger",
				"quantity": 1,
			},
			map[string]any{
				"name":     "ice-cream",
				"quantity": 1,
			},
		},
	}
	expectedOrderNumber := int64(1)
	expectedItems := []item.Item{
		{
			Name:     "hamburger",
			Quantity: 2,
		},
		{
			Name:     "cheeseburger",
			Quantity: 1,
		},
		{
			Name:     "ice-cream",
			Quantity: 1,
		},
	}
	expectedMessages := make([]*s.KitchenRequestMessage, 0)
	expectedMessages = append(expectedMessages, s.NewKitchenRequestMessage("hamburger", 2))
	expectedMessages = append(expectedMessages, s.NewKitchenRequestMessage("cheeseburger", 1))

	bodySlice, _ := json.Marshal(order)
	reqBody := bytes.NewBuffer(bodySlice)

	req, _ := http.NewRequest("POST", "/order", reqBody)
	resp := httptest.NewRecorder()

	repository := NewRepository(database)

	endpoints := NewOrderEndpoints(database, kitchenRequestsKafkaConfig, orderStatusKafkaConfig, shelf.NewEmptyShelf())
	engine := utils.SetUpRouter(endpoints.Setup)

	kitchenRequestsReader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:   kitchenRequestsKafkaConfig.Brokers,
		Topic:     topic,
		Partition: 0,
		MinBytes:  10e3, // 10KB
		MaxBytes:  10e6, // 10MB
	})

	// when
	engine.ServeHTTP(resp, req)

	// then
	assert.Equal(t, http.StatusCreated, resp.Code)

	// and
	var payload map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &payload)
	if err != nil {
		assert.Fail(t, "Error while unmarshalling response payload to map", err)
	}

	actual, exists := payload["orderNumber"]
	if !exists {
		assert.Fail(t, "Cannot determine order number from response")
	}
	actualOrderNumber := cast.ToInt64(actual)
	assert.Equal(t, expectedOrderNumber, actualOrderNumber)

	// and
	actualOrder, err := repository.FetchByOrderNumber(context.TODO(), actualOrderNumber)
	if err != nil {
		if err != nil {
			assert.Fail(t, "Failed to read order by number from DB")
		}

		assert.Equal(t, expectedOrderNumber, actualOrder.OrderNumber)
		assert.Equal(t, 10, actualOrder.CustomerId)
		assert.Equal(t, InProgress, actualOrder.Status)
		assert.Equal(t, expectedItems, actualOrder.Items)

		assert.Len(t, actualOrder.PackedItems, 1)
		assert.Equal(t, item.Item{Name: "ice-cream", Quantity: 1}, actualOrder.PackedItems[0])

		// and
		actualMessages := ReadMessages(t)

		assert.Equal(t, len(expectedMessages), len(actualMessages))
		assert.Equal(t, expectedMessages, actualMessages)

		defer func() {
			utils.TerminateKafkaReader(t, kitchenRequestsReader)
			utils.DeleteMany(t, collectionDb, bson.D{})
			utils.DeleteMany(t, orderNumberCollectionDb, bson.D{})
		}()
	}
}

func ReadMessages(t *testing.T) []*s.KitchenRequestMessage {
	retries := 2
	messages := make([]kafka.Message, 0)
	actualMessages := make([]*s.KitchenRequestMessage, 0)

	for i := 0; i < 4; i++ {
		log.Print("Reading messages from test topic", topic)
		message, err := kitchenRequestsReader.ReadMessage(context.Background())

		if err != nil {
			if retries > 0 {
				log.Println("retry reading message from broker....")
				retries--
				continue
			} else {
				assert.Fail(t, "failed reading message on test topic", err)
				return nil
			}
		}
		messages = append(messages, message)
	}

	for _, message := range messages {
		if message.Key != nil {
			actualMessage := &s.KitchenRequestMessage{}
			err := json.Unmarshal(message.Value, actualMessage)
			if err != nil {
				assert.Fail(t, "failed to unmarshal message on test topic", err)
				return nil
			}
			actualMessages = append(actualMessages, actualMessage)
		}
	}

	return actualMessages
}
