package order

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"mc-burger-orders/utils"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchOrders(t *testing.T) {
	// given
	engine := utils.SetUpRouter(Endpoints)

	req, _ := http.NewRequest("GET", "/order", nil)
	resp := httptest.NewRecorder()

	expectedItems := []interface{}{
		map[string]interface{}{
			"name":     "hamburger",
			"quantity": 1.0,
		},
		map[string]interface{}{
			"name":     "fries",
			"quantity": 1.0,
		},
	}
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

	// and
	orderItems := orderOne["items"]
	assert.Equal(t, expectedItems, orderItems)
}

func TestNewOrder_StoreOrder(t *testing.T) {
	// given
	engine := utils.SetUpRouter(Endpoints)

	order := map[string]any{
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
}

func TestNewOrder_ReturnBadRequest_WhenItemsMissing(t *testing.T) {
	// given
	engine := utils.SetUpRouter(Endpoints)

	order := map[string]any{"customerId": 10}
	bodySlice, _ := json.Marshal(order)
	reqBody := bytes.NewBuffer(bodySlice)

	req, _ := http.NewRequest("PUT", "/order", reqBody)
	resp := httptest.NewRecorder()

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
}

func TestNewOrder_ReturnBadRequest_WhenItemsEmpty(t *testing.T) {
	// given
	engine := utils.SetUpRouter(Endpoints)

	order := map[string]any{
		"customerId": 10,
		"items":      []string{},
	}
	bodySlice, _ := json.Marshal(order)
	reqBody := bytes.NewBuffer(bodySlice)

	req, _ := http.NewRequest("PUT", "/order", reqBody)
	resp := httptest.NewRecorder()

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
}
