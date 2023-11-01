package order

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
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
)

type FakeOrderEndpoints struct {
	s              *stack.Stack
	repository     m.OrderRepository
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

func TestFetchOrders(t *testing.T) {
	// given
	fakeEndpoints := FakeOrderEndpoints{
		s:              stack.NewStack(stack.CleanStack()),
		repository:     m.OrderRepository{},
		queryService:   m.OrderQueryService{Repository: m.OrderRepository{}},
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
	assert.Equal(t, 1000.0, orderOne["id"])
	assert.Equal(t, 1.0, orderOne["customerId"])
	assert.Equal(t, "READY", orderOne["status"])
}

func TestNewOrder_StoreOrder(t *testing.T) {
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
		ID:         1010,
		CustomerId: 10,
		Status:     m.Requested,
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

	fakeEndpoints := FakeOrderEndpoints{
		s:              stack.NewStack(stack.CleanStack()),
		repository:     m.OrderRepository{},
		queryService:   m.OrderQueryService{Repository: m.OrderRepository{}},
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

	assert.Equal(t, 1010.0, payload["id"])
	assert.Equal(t, 10.0, payload["customerId"])
	assert.Equal(t, "REQUESTED", payload["status"])
	assert.Equal(t, expectedItems, payload["items"])

	// and
	assert.True(t, fakeEndpoints.handler.methodCalled)
}

func TestNewOrder_ReturnBadRequest_WhenItemsMissing(t *testing.T) {
	// given
	order := map[string]any{"customerId": 10}
	bodySlice, _ := json.Marshal(order)
	reqBody := bytes.NewBuffer(bodySlice)

	req, _ := http.NewRequest("PUT", "/order", reqBody)
	resp := httptest.NewRecorder()

	fakeEndpoints := FakeOrderEndpoints{
		s:              stack.NewStack(stack.CleanStack()),
		repository:     m.OrderRepository{},
		queryService:   m.OrderQueryService{Repository: m.OrderRepository{}},
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

func TestNewOrder_ReturnBadRequest_WhenItemsEmpty(t *testing.T) {
	// given
	order := map[string]any{
		"customerId": 10,
		"items":      []string{},
	}
	bodySlice, _ := json.Marshal(order)
	reqBody := bytes.NewBuffer(bodySlice)

	req, _ := http.NewRequest("PUT", "/order", reqBody)
	resp := httptest.NewRecorder()

	fakeEndpoints := FakeOrderEndpoints{
		s:              stack.NewStack(stack.CleanStack()),
		repository:     m.OrderRepository{},
		queryService:   m.OrderQueryService{Repository: m.OrderRepository{}},
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
