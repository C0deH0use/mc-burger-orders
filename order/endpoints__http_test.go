package order

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"mc-burger-orders/command"
	"mc-burger-orders/item"
	m "mc-burger-orders/order/model"
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
)

func TestOrderHttpEndpoints(t *testing.T) {
	ctx := context.Background()
	mongoContainer = utils.TestWithMongo(ctx)
	database = utils.GetMongoDbFrom(mongoContainer)
	collectionDb = database.Collection("orders")
	orderNumberCollectionDb = database.Collection("order-numbers")

	t.Run("should return orders", shouldFetchOrdersWhenMultipleStored)
	t.Run("should store order when request is valid", shouldExecuteCommandAndStoreNewOrderWhenRequested)

	t.Cleanup(func() {
		utils.TerminateMongo(mongoContainer)
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

	endpoints := NewOrderEndpoints(database, &command.DefaultHandler{})
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
		}, {
			Name:     "ice-cream",
			Quantity: 1,
		},
	}

	bodySlice, _ := json.Marshal(order)
	reqBody := bytes.NewBuffer(bodySlice)

	req, _ := http.NewRequest("PUT", "/order", reqBody)
	resp := httptest.NewRecorder()

	repository := m.NewRepository(database)

	endpoints := NewOrderEndpoints(database, &command.DefaultHandler{})
	engine := utils.SetUpRouter(endpoints.Setup)

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

	defer func() {
		utils.DeleteMany(collectionDb, bson.D{})
		utils.DeleteMany(orderNumberCollectionDb, bson.D{})
	}()
}
