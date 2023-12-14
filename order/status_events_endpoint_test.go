package order

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"mc-burger-orders/event"
	"mc-burger-orders/kitchen/item"
	"mc-burger-orders/testing/utils"
	"net/http"
	"strconv"
	"strings"
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
	expectedOrders := []interface{}{
		Order{OrderNumber: 1000, CustomerId: 3, Items: []item.Item{{Name: "cheeseburger", Quantity: 2}, {Name: "cheeseburger", Quantity: 3}}, Status: Requested, CreatedAt: time.Now(), ModifiedAt: time.Now()},
	}
	utils.DeleteMany(t, collectionDb, bson.D{})
	utils.InsertMany(t, collectionDb, expectedOrders)

	endpoints := NewOrderStatusEventsEndpoints(database, orderStatusKafkaConfig)
	engine := utils.SetUpRouter(endpoints.Setup)
	go runEngine(engine, t)

	req, _ := http.NewRequest("GET", "http://localhost:8080/order/status", nil)

	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Connection", "keep-alive")

	sendStatusUpdateMessages(t, InProgress)
	sendStatusUpdateMessages(t, Ready)

	// when
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		assert.Fail(t, err.Error())
		return
	}

	// then
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// and
	actualResponses := make([]map[string]interface{}, 0)
	for {
		body := make([]byte, 80)
		var payload map[string]interface{}
		_, err := resp.Body.Read(body)
		if err != nil {
			assert.Fail(t, "Error while ready bytes from response", err)
		}

		bodyStr := string(body)
		bodyStr = strings.Replace(bodyStr, "event:order-status\ndata:", "", 1)
		bodyStr = strings.SplitAfter(bodyStr, "}")[0]
		err = json.Unmarshal([]byte(bodyStr), &payload)
		if err != nil {
			assert.Fail(t, "Error while unmarshalling response payload to map", err)
		}
		actualResponses = append(actualResponses, payload)

		if len(actualResponses) >= 2 {
			break
		}
	}

	assert.Equal(t, float64(1000), actualResponses[0]["orderNumber"])
	assert.Equal(t, float64(1000), actualResponses[1]["orderNumber"])

	assert.Equal(t, "IN_PROGRESS", actualResponses[0]["status"])
	assert.Equal(t, "READY", actualResponses[1]["status"])
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

func runEngine(engine *gin.Engine, t *testing.T) {
	err := engine.Run()
	if err != nil {
		assert.Fail(t, err.Error())
		return
	}
}
