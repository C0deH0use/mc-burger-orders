package order

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"mc-burger-orders/command"
	"mc-burger-orders/item"
	m "mc-burger-orders/order/model"
	s "mc-burger-orders/order/service"
	"mc-burger-orders/utils"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

var (
	mongoContainer          *mongodb.MongoDBContainer
	database                *mongo.Database
	collectionDb            *mongo.Collection
	orderNumberCollectionDb *mongo.Collection
	kafkaConfig             s.KitchenServiceConfigs
	testReader              *kafka.Reader
	topic                   = "test-kitchen-requests"
)

func TestOrderHttpEndpoints(t *testing.T) {
	ctx := context.Background()
	mongoContainer = utils.TestWithMongo(ctx)
	kafkaContainer := utils.TestWithKafka(ctx)
	brokers, err := kafkaContainer.Brokers(ctx)
	if err != nil {
		assert.Fail(t, "cannot read Brokers from kafka container")
	}
	kafkaConfig = s.KitchenServiceConfigs{
		Topic:             topic,
		Brokers:           brokers,
		NumPartitions:     1,
		ReplicationFactor: 1,
	}
	database = utils.GetMongoDbFrom(mongoContainer)
	collectionDb = database.Collection("orders")
	orderNumberCollectionDb = database.Collection("order-numbers")
	t.Run("should return orders", shouldFetchOrdersWhenMultipleStored)
	t.Run("should store order when request is valid", shouldExecuteCommandAndStoreNewOrderWhenRequested)

	t.Cleanup(func() {
		log.Println("Running Clean UP code")
		utils.TerminateMongo(mongoContainer)
		//utils.TerminateKafka(kafkaContainer)
	})
}

func shouldFetchOrdersWhenMultipleStored(t *testing.T) {
	// given
	expectedOrders := []interface{}{
		m.Order{OrderNumber: 1000, CustomerId: 1, Items: []item.Item{{Name: "hamburger", Quantity: 1}, {Name: "fries", Quantity: 1}}, Status: m.Ready, CreatedAt: time.Now(), ModifiedAt: time.Now()},
		m.Order{OrderNumber: 1001, CustomerId: 2, Items: []item.Item{{Name: "hamburger", Quantity: 1}, {Name: "cheeseburger", Quantity: 2}}, Status: m.InProgress, CreatedAt: time.Now(), ModifiedAt: time.Now()},
		m.Order{OrderNumber: 1002, CustomerId: 3, Items: []item.Item{{Name: "cheeseburger", Quantity: 2}, {Name: "cheeseburger", Quantity: 3}}, Status: m.Requested, CreatedAt: time.Now(), ModifiedAt: time.Now()},
	}
	utils.DeleteMany(collectionDb, bson.D{})
	utils.InsertMany(collectionDb, expectedOrders)

	endpoints := NewOrderEndpoints(database, kafkaConfig, &command.DefaultHandler{})
	engine := utils.SetUpRouter(endpoints.Setup)

	req, _ := http.NewRequest("GET", "/order", nil)
	resp := httptest.NewRecorder()

	// when
	engine.ServeHTTP(resp, req)

	// then
	assert.Equal(t, http.StatusOK, resp.Code)

	// and
	var payload []map[string]any
	err := json.Unmarshal([]byte(resp.Body.String()), &payload)
	if err != nil {
		assert.Fail(t, "Error while unmarshalling response payload to map", err)
	}

	assert.Equal(t, 3, len(payload))

	orderOne := payload[0]
	assert.Equal(t, 1000.0, orderOne["orderNumber"])
	assert.Equal(t, 1.0, orderOne["customerId"])
	assert.Equal(t, "READY", orderOne["status"])

	defer func() {
		utils.DeleteMany(collectionDb, bson.D{})
		utils.DeleteMany(orderNumberCollectionDb, bson.D{})
	}()
}

func shouldExecuteCommandAndStoreNewOrderWhenRequested(t *testing.T) {
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
	expectedMessages := make([]*s.KitchenRequestMessage, 1)
	expectedMessages = append(expectedMessages, s.NewKitchenRequestMessage("hamburger", 2))
	expectedMessages = append(expectedMessages, s.NewKitchenRequestMessage("cheeseburger", 1))

	bodySlice, _ := json.Marshal(order)
	reqBody := bytes.NewBuffer(bodySlice)

	req, _ := http.NewRequest("PUT", "/order", reqBody)
	resp := httptest.NewRecorder()

	repository := m.NewRepository(database)

	endpoints := NewOrderEndpoints(database, kafkaConfig, &command.DefaultHandler{})
	engine := utils.SetUpRouter(endpoints.Setup)

	testReader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:   kafkaConfig.Brokers,
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
	err := json.Unmarshal([]byte(resp.Body.String()), &payload)
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
		assert.Fail(t, "Failed to read order by number from DB")
	}

	assert.Equal(t, expectedOrderNumber, actualOrder.OrderNumber)
	assert.Equal(t, 10, actualOrder.CustomerId)
	assert.Equal(t, m.InProgress, actualOrder.Status)
	assert.Equal(t, expectedItems, actualOrder.Items)

	assert.Len(t, actualOrder.PackedItems, 1)
	assert.Equal(t, item.Item{Name: "ice-cream", Quantity: 1}, actualOrder.PackedItems[0])

	// and
	receivedMessages := make([]kafka.Message, 10)
	retries := 2
	for i := 0; i < retries; i++ {
		message, err := testReader.ReadMessage(context.Background())

		if err != nil {
			log.Print(t, "failed reading message on test Topic", err)
			break
		}
		receivedMessages = append(receivedMessages, message)
	}

	// and
	actualMessages := make([]*s.KitchenRequestMessage, 10)

	for _, message := range receivedMessages {
		actualMessage := &s.KitchenRequestMessage{}
		err = json.Unmarshal(message.Value, actualMessage)
		if err != nil {
			assert.Fail(t, "failed to unmarshal message on test Topic", err)
		}
		actualMessages = append(actualMessages, actualMessage)
	}

	assert.Equal(t, expectedMessages, actualMessages)

	defer func() {
		utils.TerminateKafkaReader(testReader)
		utils.DeleteMany(collectionDb, bson.D{})
		utils.DeleteMany(orderNumberCollectionDb, bson.D{})
	}()
}
