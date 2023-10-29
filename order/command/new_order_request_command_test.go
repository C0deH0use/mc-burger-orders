package command

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"io"
	"mc-burger-orders/order/model"
	"mc-burger-orders/stack"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewOrderRequestCommand_CreateNewOrder(t *testing.T) {
	// given
	repository := model.OrderRepository{}
	sk := stack.CleanStack()
	sk["hamburger"] = 3

	s := stack.NewStack(sk)
	resp := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(resp)

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
	c.Request = &http.Request{
		Header: make(http.Header),
		Body:   io.NopCloser(bytes.NewBuffer(bodySlice)),
	}

	command := &NewOrderRequestCommand{
		Repository: repository,
		Stack:      s,
	}

	// when
	command.CreateNewOrder(c)

	// then
	var payload map[string]any
	err := json.Unmarshal([]byte(resp.Body.String()), &payload)
	if err != nil {
		fmt.Println("Error while unmarshalling response payload to map", err)
	}

	assert.Equal(t, 1010.0, payload["id"])
	assert.Equal(t, 10.0, payload["customerId"])
	assert.Equal(t, "REQUESTED", payload["status"])

	actualItems := payload["items"].([]interface{})
	assert.Equal(t, 2, len(actualItems))

	expectedPackedItems := []interface{}{
		map[string]interface{}{
			"name":     "hamburger",
			"quantity": 2.0,
		},
		map[string]interface{}{
			"name":     "ice-cream",
			"quantity": 1.0,
		},
	}
	actualPackedItems := payload["packedItems"].([]interface{})
	assert.Equal(t, 2, len(actualPackedItems))
	assert.Equal(t, expectedPackedItems, actualPackedItems)

	// and
	assert.Equal(t, 1, s.GetCurrent("hamburger"))
}

func TestNewOrderRequestCommand_CreateNewOrderAndPackOnlyTheseItemsThatAreAvailable(t *testing.T) {
	// given
	repository := model.OrderRepository{}
	sk := stack.CleanStack()
	sk["hamburger"] = 1

	s := stack.NewStack(sk)
	resp := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(resp)

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
	c.Request = &http.Request{
		Header: make(http.Header),
		Body:   io.NopCloser(bytes.NewBuffer(bodySlice)),
	}

	command := &NewOrderRequestCommand{
		Repository: repository,
		Stack:      s,
	}

	// when
	command.CreateNewOrder(c)

	// then
	var payload map[string]any
	err := json.Unmarshal([]byte(resp.Body.String()), &payload)
	if err != nil {
		fmt.Println("Error while unmarshalling response payload to map", err)
	}

	assert.Equal(t, 1010.0, payload["id"])
	assert.Equal(t, 10.0, payload["customerId"])
	assert.Equal(t, "REQUESTED", payload["status"])

	actualItems := payload["items"].([]interface{})
	assert.Equal(t, 2, len(actualItems))

	expectedPackedItems := []interface{}{
		map[string]interface{}{
			"name":     "hamburger",
			"quantity": 1.0,
		},
		map[string]interface{}{
			"name":     "ice-cream",
			"quantity": 1.0,
		},
	}
	actualPackedItems := payload["packedItems"].([]interface{})
	assert.Equal(t, 2, len(actualPackedItems))
	assert.Equal(t, expectedPackedItems, actualPackedItems)

	// and
	assert.Equal(t, 0, s.GetCurrent("hamburger"))
}
