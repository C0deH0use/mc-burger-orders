package order

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"mc-burger-orders/item"
	"mc-burger-orders/middleware"
	command2 "mc-burger-orders/order/command"
	m "mc-burger-orders/order/model"
	"mc-burger-orders/order/service"
	"mc-burger-orders/stack"
	"mc-burger-orders/utils"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type FakeOrderEndpoints struct {
	s              *stack.Stack
	repository     *m.OrderRepository
	queryService   m.OrderQueryService
	kitchenService service.KitchenRequestService
	handler        *FakeCommandHandler
}

type FakeCommandHandler struct {
	order        m.Order
	methodCalled bool
}

func (e *FakeCommandHandler) Execute(c command2.Command) (*m.Order, error) {
	e.methodCalled = true
	return &e.order, nil
}

func (f *FakeOrderEndpoints) FakeEndpoints() middleware.EndpointsSetup {
	return &Endpoints{
		stack:          f.s,
		queryService:   f.queryService,
		repository:     f.repository,
		kitchenService: f.kitchenService,
		commandHandler: f.handler,
	}
}

var mongoContainer *mongodb.MongoDBContainer
var database *mongo.Database

func TestOrderEndpoints(t *testing.T) {
	mongoContainer = utils.TestWithMongo(t)
	database = utils.GetMongoDbFrom(mongoContainer)

	t.Run("should return orders", fetchOrders)
	t.Run("should store order when request is valid", shouldStoreNewOrder)
	t.Run("should return BAD REQUEST when request has no items", shouldReturnBadRequestWhenItemsEmpty)
	t.Run("should return BAD REQUEST when request is missing items", shouldReturnBadRequestWhenNoItems)

	t.Cleanup(func() {
		utils.TerminateMongo(mongoContainer)
	})
}

func fetchOrders(t *testing.T) {
	// given
	collectionDb := database.Collection("orders")
	expectedOrders := []interface{}{
		m.Order{OrderNumber: 1000, CustomerId: 1, Items: []item.Item{{Name: "hamburger", Quantity: 1}, {Name: "fries", Quantity: 1}}, Status: m.Ready, CreatedAt: time.Now(), ModifiedAt: time.Now()},
		m.Order{OrderNumber: 1001, CustomerId: 2, Items: []item.Item{{Name: "hamburger", Quantity: 1}, {Name: "cheeseburger", Quantity: 2}}, Status: m.InProgress, CreatedAt: time.Now(), ModifiedAt: time.Now()},
		m.Order{OrderNumber: 1002, CustomerId: 3, Items: []item.Item{{Name: "cheeseburger", Quantity: 2}, {Name: "cheeseburger", Quantity: 3}}, Status: m.Requested, CreatedAt: time.Now(), ModifiedAt: time.Now()},
	}
	utils.DeleteMany(collectionDb, bson.D{})
	utils.InsertMany(collectionDb, expectedOrders)

	repository := m.NewRepository(database)
	fakeEndpoints := FakeOrderEndpoints{
		s:              stack.NewStack(stack.CleanStack()),
		repository:     repository,
		queryService:   m.OrderQueryService{Repository: repository},
		kitchenService: &service.KitchenService{},
	}
	endpoints := fakeEndpoints.FakeEndpoints()
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
		fmt.Println("Error while unmarshalling response payload to map", err)
	}

	assert.Equal(t, 3, len(payload))

	orderOne := payload[0]
	assert.Equal(t, 1000.0, orderOne["orderNumber"])
	assert.Equal(t, 1.0, orderOne["customerId"])
	assert.Equal(t, "READY", orderOne["status"])

	defer func() {
		utils.DeleteMany(collectionDb, bson.D{})
	}()
}

func shouldStoreNewOrder(t *testing.T) {
	// given
	collectionDb := database.Collection("orders")

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
	bodySlice, _ := json.Marshal(order)
	reqBody := bytes.NewBuffer(bodySlice)

	req, _ := http.NewRequest("PUT", "/order", reqBody)
	resp := httptest.NewRecorder()
	expectedItems := []interface{}{
		map[string]interface{}{
			"name":     "hamburger",
			"quantity": 2.0,
		},
		map[string]interface{}{
			"name":     "ice-cream",
			"quantity": 1.0,
		},
	}
	expectedOrder := m.Order{
		OrderNumber: 1010,
		CustomerId:  10,
		Status:      m.Requested,
		Items: []item.Item{
			{
				Name:     "hamburger",
				Quantity: 2,
			}, {
				Name:     "ice-cream",
				Quantity: 1,
			},
		},
	}
	repository := m.NewRepository(database)
	fakeEndpoints := FakeOrderEndpoints{
		s:              stack.NewStack(stack.CleanStack()),
		repository:     repository,
		queryService:   m.OrderQueryService{Repository: repository},
		kitchenService: &service.KitchenService{},
		handler:        &FakeCommandHandler{order: expectedOrder},
	}
	endpoints := fakeEndpoints.FakeEndpoints()
	engine := utils.SetUpRouter(endpoints.Setup)

	// when
	engine.ServeHTTP(resp, req)

	// then
	assert.Equal(t, http.StatusCreated, resp.Code)

	// and
	var payload map[string]any
	err := json.Unmarshal([]byte(resp.Body.String()), &payload)
	if err != nil {
		fmt.Println("Error while unmarshalling response payload to map", err)
	}

	assert.Equal(t, 1010.0, payload["orderNumber"])
	assert.Equal(t, 10.0, payload["customerId"])
	assert.Equal(t, "REQUESTED", payload["status"])
	assert.Equal(t, expectedItems, payload["items"])

	// and
	assert.True(t, fakeEndpoints.handler.methodCalled)

	defer func() {
		utils.DeleteMany(collectionDb, bson.D{})
	}()
}

func shouldReturnBadRequestWhenNoItems(t *testing.T) {
	// given
	order := map[string]any{"customerId": 10}
	bodySlice, _ := json.Marshal(order)
	reqBody := bytes.NewBuffer(bodySlice)

	req, _ := http.NewRequest("PUT", "/order", reqBody)
	resp := httptest.NewRecorder()

	repository := m.NewRepository(database)
	fakeEndpoints := FakeOrderEndpoints{
		s:              stack.NewStack(stack.CleanStack()),
		repository:     repository,
		queryService:   m.OrderQueryService{Repository: repository},
		kitchenService: &service.KitchenService{},
		handler:        &FakeCommandHandler{},
	}
	endpoints := fakeEndpoints.FakeEndpoints()
	engine := utils.SetUpRouter(endpoints.Setup)

	// when
	engine.ServeHTTP(resp, req)

	// then
	assert.Equal(t, http.StatusBadRequest, resp.Code)

	// and
	var payload map[string]any
	err := json.Unmarshal([]byte(resp.Body.String()), &payload)
	if err != nil {
		fmt.Println("Error while unmarshalling response payload to map", err)
	}

	assert.Equal(t, "Schema Error. Key: 'NewOrder.Items' Error:Field validation for 'Items' failed on the 'required' tag", payload["errorMessage"])

	// and
	assert.False(t, fakeEndpoints.handler.methodCalled)
}

func shouldReturnBadRequestWhenItemsEmpty(t *testing.T) {
	// given
	order := map[string]any{
		"customerId": 10,
		"items":      []string{},
	}
	bodySlice, _ := json.Marshal(order)
	reqBody := bytes.NewBuffer(bodySlice)

	req, _ := http.NewRequest("PUT", "/order", reqBody)
	resp := httptest.NewRecorder()

	repository := m.NewRepository(database)
	fakeEndpoints := FakeOrderEndpoints{
		s:              stack.NewStack(stack.CleanStack()),
		repository:     repository,
		queryService:   m.OrderQueryService{Repository: repository},
		kitchenService: &service.KitchenService{},
		handler:        &FakeCommandHandler{},
	}
	endpoints := fakeEndpoints.FakeEndpoints()
	engine := utils.SetUpRouter(endpoints.Setup)

	// when
	engine.ServeHTTP(resp, req)

	// then
	assert.Equal(t, http.StatusBadRequest, resp.Code)

	// and
	var payload map[string]any
	err := json.Unmarshal([]byte(resp.Body.String()), &payload)
	if err != nil {
		fmt.Println("Error while unmarshalling response payload to map", err)
	}

	assert.Equal(t, "Schema Error. Key: 'NewOrder.Items' Error:Field validation for 'Items' failed on the 'gt' tag", payload["errorMessage"])

	// and
	assert.False(t, fakeEndpoints.handler.methodCalled)
}
