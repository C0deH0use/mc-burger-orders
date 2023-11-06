package order

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"
	command2 "mc-burger-orders/command"
	"mc-burger-orders/middleware"
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
	result       bool
	methodCalled bool
}

type FakeRepository struct {
	orders       []*m.Order
	methodCalled bool
}

func (e *FakeCommandHandler) Execute(c command2.Command) (bool, error) {
	e.methodCalled = true
	return e.result, nil
}

func (f *FakeOrderEndpoints) FakeEndpoints() middleware.EndpointsSetup {
	return &Endpoints{
		stack:           f.s,
		queryService:    f.queryService,
		orderRepository: f.repository,
		kitchenService:  f.kitchenService,
		commandHandler:  f.handler,
	}
}

func (e *FakeRepository) InsertOrUpdate(ctx context.Context, order m.Order) (*m.Order, error) {
	e.methodCalled = true
	return e.orders[0], nil
}

func (e *FakeRepository) FetchById(ctx context.Context, id interface{}) (*m.Order, error) {
	e.methodCalled = true
	return e.orders[0], nil
}
func (e *FakeRepository) FetchMany(ctx context.Context) ([]m.Order, error) {
	e.methodCalled = true
	orders := make([]m.Order, len(e.orders))
	for _, valPointer := range e.orders {
		orders = append(orders, *valPointer)
	}
	return orders, nil
}

func (e *FakeRepository) GetNext(ctx context.Context) (int64, error) {
	return expectedOrderNumber, nil
}

var expectedOrderNumber = int64(10)

func Test_UnitOrderEndpoints(t *testing.T) {
	t.Run("should execute order request command when request is valid", shouldExecuteNewOrderCommand)
	t.Run("should return BAD REQUEST when request has no items", shouldReturnBadRequestWhenItemsEmpty)
	t.Run("should return BAD REQUEST when request is missing items", shouldReturnBadRequestWhenNoItems)
}

func shouldExecuteNewOrderCommand(t *testing.T) {
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

	repository := &FakeRepository{}
	orderNumberRepository := &FakeRepository{}

	fakeEndpoints := FakeOrderEndpoints{
		s:              stack.NewStack(stack.CleanStack()),
		repository:     repository,
		queryService:   m.OrderQueryService{Repository: repository, OrderNumberRepository: orderNumberRepository},
		kitchenService: &service.KitchenService{},
		handler:        &FakeCommandHandler{result: true},
	}
	endpoints := fakeEndpoints.FakeEndpoints()
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
	assert.True(t, fakeEndpoints.handler.methodCalled)
}

func shouldReturnBadRequestWhenNoItems(t *testing.T) {
	// given
	order := map[string]any{"customerId": 10}
	bodySlice, _ := json.Marshal(order)
	reqBody := bytes.NewBuffer(bodySlice)

	req, _ := http.NewRequest("PUT", "/order", reqBody)
	resp := httptest.NewRecorder()

	repository := &FakeRepository{}
	orderNumberRepository := &FakeRepository{}

	fakeEndpoints := FakeOrderEndpoints{
		s:              stack.NewStack(stack.CleanStack()),
		repository:     repository,
		queryService:   m.OrderQueryService{Repository: repository, OrderNumberRepository: orderNumberRepository},
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
		assert.Fail(t, "Error while unmarshalling response payload to map", err)
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

	repository := &FakeRepository{}
	orderNumberRepository := &FakeRepository{}

	fakeEndpoints := FakeOrderEndpoints{
		s:              stack.NewStack(stack.CleanStack()),
		repository:     repository,
		queryService:   m.OrderQueryService{Repository: repository, OrderNumberRepository: orderNumberRepository},
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
		assert.Fail(t, "Error while unmarshalling response payload to map", err)
	}

	assert.Equal(t, "Schema Error. Key: 'NewOrder.Items' Error:Field validation for 'Items' failed on the 'gt' tag", payload["errorMessage"])

	// and
	assert.False(t, fakeEndpoints.handler.methodCalled)
}
