package order

import (
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"mc-burger-orders/event"
	"mc-burger-orders/kitchen/item"
	"mc-burger-orders/testing/utils"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

func TestIntegrationOrderStatus_HttpEndpoints(t *testing.T) {
	utils.IntegrationTest(t)
	ctx := context.Background()
	mongoContainer, database = utils.TestWithMongo(t, ctx)
	kafkaContainer, brokers := utils.TestWithKafka(t, ctx)
	collectionDb = database.Collection("orders")

	orderStatusKafkaConfig = event.TestTopicConfigs(orderStatusTopic, brokers...)

	t.Run("should get order updates when order changes", shouldGetOrderUpdatesWhenOrderChanges)

	t.Cleanup(func() {
		t.Log("Running Clean UP code")
		utils.TerminateMongo(t, ctx, mongoContainer)
		utils.TerminateKafka(t, ctx, kafkaContainer)
	})
}

func shouldGetOrderUpdatesWhenOrderChanges(t *testing.T) {
	// given
	orderStatusChannel := make(chan kafka.Message)
	expectedOrders := []interface{}{
		Order{OrderNumber: 1000, CustomerId: 3, Items: []item.Item{{Name: "cheeseburger", Quantity: 2}, {Name: "cheeseburger", Quantity: 3}}, Status: Requested, CreatedAt: time.Now(), ModifiedAt: time.Now()},
	}
	utils.DeleteMany(t, collectionDb, bson.D{})
	utils.InsertMany(t, collectionDb, expectedOrders)

	endpoints := NewOrderStatusEventsEndpoints(database, orderStatusKafkaConfig)
	engine := utils.SetUpRouter(endpoints.Setup)

	req, _ := http.NewRequest("GET", "/order/status", nil)
	resp := httptest.NewRecorder()

	sendStatusUpdateMessages(t, InProgress)
	sendStatusUpdateMessages(t, Ready)

	// when
	engine.ServeHTTP(resp, req)

	// then
	assert.Equal(t, http.StatusOK, resp.Code)

	// and
	var payload map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &payload)
	if err != nil {
		assert.Fail(t, "Error while unmarshalling response payload to map", err)
	}

	actualOrderNumber := payload["orderNumber"]
	actualStatus := payload["status"]

	assert.Equal(t, "1000", actualOrderNumber)
	assert.Equal(t, InProgress, actualStatus)

	defer close(orderStatusChannel)
}

func sendStatusUpdateMessages(t *testing.T, status OrderStatus) {
	writer := event.NewTopicWriter(orderStatusKafkaConfig)

	msgKey := []byte(strconv.FormatInt(time.Now().UnixNano(), 10))

	headers := make([]kafka.Header, 0)
	headers = append(headers, kafka.Header{Key: "order", Value: []byte("1000")})
	headers = append(headers, kafka.Header{Key: "event", Value: []byte(StatusUpdatedEvent)})

	requestPayload := map[string]any{
		"status": status,
	}
	b, _ := json.Marshal(requestPayload)

	msg := kafka.Message{
		Key:     msgKey,
		Headers: headers,
		Value:   b,
	}

	// when
	err := writer.SendMessage(context.Background(), msg)
	if err != nil {
		assert.Fail(t, "failed to send message to topic", stackKafkaConfig.Topic)
	}
}
