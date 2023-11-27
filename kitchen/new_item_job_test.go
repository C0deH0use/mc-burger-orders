package kitchen

import (
	"encoding/json"
	"fmt"
	"github.com/gammazero/workerpool"
	"github.com/segmentio/kafka-go"
	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"
	"mc-burger-orders/stack"
	"mc-burger-orders/testing/data"
	"testing"
	"time"
)

var (
	expectedOrderNumber = int64(1010)
)

func TestHandler_CreateNewItem(t *testing.T) {
	t.Run("should create new items when valid requested in the message", shouldPrepareNewItemsWhenRequestedInTheMessage)
	t.Run("should create new items when missing order in the message", shouldPrepareItemsWhenMessageMissingOrderNumber)
	t.Run("should skip creation when no requests in the message", shouldSkipWhenMessageHasZeroRequests)
}

func shouldPrepareNewItemsWhenRequestedInTheMessage(t *testing.T) {
	// given
	emptyStack := stack.NewEmptyStack()
	prepMealStub := NewMealPrepService()
	handler := &Handler{
		kitchenCooks:    workerpool.New(1),
		mealPreparation: prepMealStub,
		stack:           emptyStack,
	}

	messageValue := make([]map[string]any, 0)
	messageValue = data.AppendHamburgerItem(messageValue, 1)
	messageValue = data.AppendCheeseBurgerItem(messageValue, 2)

	message := givenKafkaMessage(t, expectedOrderNumber, messageValue)

	// when
	result, err := handler.CreateNewItem(message)

	// then
	assert.True(t, result)
	assert.Nil(t, err)

	// and
	assert.True(t, prepMealStub.HaveBeenCalledWith(MealPrepMatchingFnc("hamburger", 1)))
	assert.True(t, prepMealStub.HaveBeenCalledWith(MealPrepMatchingFnc("cheeseburger", 2)))
}

func shouldPrepareItemsWhenMessageMissingOrderNumber(t *testing.T) {
	// given
	emptyStack := stack.NewEmptyStack()
	prepMealStub := NewMealPrepService()
	handler := &Handler{
		kitchenCooks:    workerpool.New(1),
		mealPreparation: prepMealStub,
		stack:           emptyStack,
	}

	messageValue := make([]map[string]any, 0)
	messageValue = data.AppendHamburgerItem(messageValue, 1)
	messageValue = data.AppendCheeseBurgerItem(messageValue, 2)

	message := givenKafkaMessage(t, -1, messageValue)

	// when
	result, err := handler.CreateNewItem(message)

	// then
	assert.True(t, result)
	assert.Nil(t, err)

	// and
	assert.True(t, prepMealStub.HaveBeenCalledWith(MealPrepMatchingFnc("hamburger", 1)))
	assert.True(t, prepMealStub.HaveBeenCalledWith(MealPrepMatchingFnc("cheeseburger", 2)))
}

func shouldSkipWhenMessageHasZeroRequests(t *testing.T) {
	// given
	emptyStack := stack.NewEmptyStack()
	prepMealStub := NewMealPrepService()
	handler := &Handler{
		kitchenCooks:    workerpool.New(1),
		mealPreparation: prepMealStub,
		stack:           emptyStack,
	}

	message := givenKafkaMessage(t, expectedOrderNumber, make([]map[string]any, 0))

	// when
	result, err := handler.CreateNewItem(message)

	// then
	assert.True(t, result)
	assert.Nil(t, err)

	// and
	assert.Equal(t, 0, prepMealStub.CalledCnt())
}

func givenKafkaMessage(t *testing.T, orderNumber int64, messageValue []map[string]any) kafka.Message {
	b, err := json.Marshal(messageValue)
	if err != nil {
		assert.Fail(t, "Could not marshal Kafka message", err)
	}
	headers := make([]kafka.Header, 0)
	if orderNumber > 0 {
		headers = append(headers, kafka.Header{Key: "order", Value: []byte(cast.ToString(expectedOrderNumber))})
	}

	headers = append(headers, kafka.Header{Key: "event", Value: []byte("request-item")})

	message := kafka.Message{
		Headers: headers,
		Topic:   "some-kafka-topic",
		Key:     []byte(fmt.Sprintf("%d", time.Now().UnixNano())),
		Value:   b,
	}
	return message
}
