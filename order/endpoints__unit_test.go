package order

import (
	"bytes"
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"
	command2 "mc-burger-orders/command"
	"mc-burger-orders/middleware"
	"mc-burger-orders/shelf"
	"mc-burger-orders/testing/utils"
	"net/http"
	"net/http/httptest"
	"testing"
)

type FakeOrderEndpoints struct {
	s              *shelf.Shelf
	repository     OrderRepository
	queryService   OrderQueryService
	kitchenService KitchenRequestService
	dispatcher     *FakeCommandDispatcher
}

type FakeCommandDispatcher struct {
	result       bool
	methodCalled bool
}

func (e *FakeCommandDispatcher) Execute(c command2.Command, message kafka.Message, commandResults chan command2.TypedResult) {
	e.methodCalled = true
	commandResults <- command2.TypedResult{Result: e.result, Type: "FakeCommandDispatcher"}
}

func (f *FakeOrderEndpoints) FakeEndpoints() middleware.EndpointsSetup {
	return &Endpoints{
		stack:           f.s,
		queryService:    f.queryService,
		orderRepository: f.repository,
		kitchenService:  f.kitchenService,
		dispatcher:      f.dispatcher,
	}
}

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

	req, _ := http.NewRequest("POST", "/order", reqBody)
	resp := httptest.NewRecorder()

	repository := GivenRepository()
	orderNumberRepository := GivenRepository()
	orderNumberRepository.ReturnNextNumber(expectedOrderNumber)

	fakeEndpoints := FakeOrderEndpoints{
		s:              shelf.NewEmptyShelf(),
		repository:     repository,
		queryService:   OrderQueryService{Repository: repository, orderNumberRepository: orderNumberRepository},
		kitchenService: &KitchenService{},
		dispatcher:     &FakeCommandDispatcher{result: true},
	}
	endpoints := fakeEndpoints.FakeEndpoints()
	engine := utils.SetUpRouter(endpoints.Setup)

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
	assert.True(t, fakeEndpoints.dispatcher.methodCalled)
}

func shouldReturnBadRequestWhenNoItems(t *testing.T) {
	// given
	order := map[string]any{"customerId": 10}
	bodySlice, _ := json.Marshal(order)
	reqBody := bytes.NewBuffer(bodySlice)

	req, _ := http.NewRequest("POST", "/order", reqBody)
	resp := httptest.NewRecorder()

	repository := GivenRepository()
	orderNumberRepository := GivenRepository()
	orderNumberRepository.ReturnNextNumber(expectedOrderNumber)

	fakeEndpoints := FakeOrderEndpoints{
		s:              shelf.NewEmptyShelf(),
		repository:     repository,
		queryService:   OrderQueryService{Repository: repository, orderNumberRepository: orderNumberRepository},
		kitchenService: &KitchenService{},
		dispatcher:     &FakeCommandDispatcher{},
	}
	endpoints := fakeEndpoints.FakeEndpoints()
	engine := utils.SetUpRouter(endpoints.Setup)

	// when
	engine.ServeHTTP(resp, req)

	// then
	assert.Equal(t, http.StatusBadRequest, resp.Code)

	// and
	var payload map[string]any
	err := json.Unmarshal(resp.Body.Bytes(), &payload)
	if err != nil {
		assert.Fail(t, "Error while unmarshalling response payload to map", err)
	}

	assert.Equal(t, "Schema Error. Key: 'NewOrder.Items' Error:Field validation for 'Items' failed on the 'required' tag", payload["errorMessage"])

	// and
	assert.False(t, fakeEndpoints.dispatcher.methodCalled)
}

func shouldReturnBadRequestWhenItemsEmpty(t *testing.T) {
	// given
	order := map[string]any{
		"customerId": 10,
		"items":      []string{},
	}
	bodySlice, _ := json.Marshal(order)
	reqBody := bytes.NewBuffer(bodySlice)

	req, _ := http.NewRequest("POST", "/order", reqBody)
	resp := httptest.NewRecorder()

	repository := GivenRepository()
	orderNumberRepository := GivenRepository()
	orderNumberRepository.ReturnNextNumber(expectedOrderNumber)

	fakeEndpoints := FakeOrderEndpoints{
		s:              shelf.NewEmptyShelf(),
		repository:     repository,
		queryService:   OrderQueryService{Repository: repository, orderNumberRepository: orderNumberRepository},
		kitchenService: &KitchenService{},
		dispatcher:     &FakeCommandDispatcher{},
	}
	endpoints := fakeEndpoints.FakeEndpoints()
	engine := utils.SetUpRouter(endpoints.Setup)

	// when
	engine.ServeHTTP(resp, req)

	// then
	assert.Equal(t, http.StatusBadRequest, resp.Code)

	// and
	var payload map[string]any
	err := json.Unmarshal(resp.Body.Bytes(), &payload)
	if err != nil {
		assert.Fail(t, "Error while unmarshalling response payload to map", err)
	}

	assert.Equal(t, "Schema Error. Key: 'NewOrder.Items' Error:Field validation for 'Items' failed on the 'gt' tag", payload["errorMessage"])

	// and
	assert.False(t, fakeEndpoints.dispatcher.methodCalled)
}
