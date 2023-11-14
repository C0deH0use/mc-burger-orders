package command

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	i "mc-burger-orders/item"
	m "mc-burger-orders/order/model"
	"mc-burger-orders/stack"
	stubs2 "mc-burger-orders/testing/stubs"
	"testing"
)

var (
	hamburger           = "hamburger"
	cheeseburger        = "cheeseburger"
	mcSpicy             = "mc-spicy"
	spicyStripes        = "spicy-stripes"
	expectedOrderNumber = int64(1011)
	orderItems          = []i.Item{
		{Name: hamburger, Quantity: 2},
		{Name: cheeseburger, Quantity: 2},
		{Name: mcSpicy, Quantity: 3},
		{Name: spicyStripes, Quantity: 8},
	}
)

func TestPackItemCommand_Execute(t *testing.T) {

	t.Run("should add new items added to stack when command executed", shouldPackItemPointedInMessage)
	t.Run("should set to READY when all items are packed", shouldFinishPackingOrderWhenLastItemsCameFromKitchen)
	t.Run("should pack available items and request new when not all items are available", shouldRequestAdditionalItemWhenMoreAreNeeded)
	t.Run("should fail when message key is missing order number", shouldFailWhenMessageKeyMissingOrderNumber)
	t.Run("should fail when message key has invalid order number", shouldFailWhenMessageKeyHasInvalidOrderNumber)
	t.Run("should fail when message value is empty", shouldFailWhenMessageValueIsEmpty)
}

func shouldPackItemPointedInMessage(t *testing.T) {
	// given
	s := stack.NewStack(stack.CleanStack())
	s.AddMany(hamburger, 3)
	s.AddMany(cheeseburger, 2)
	s.AddMany(mcSpicy, 4)
	s.AddMany(spicyStripes, 2)

	messageValue := make([]map[string]any, 0)
	messageValue = append(messageValue, map[string]any{
		"itemName": mcSpicy,
		"quantity": 2,
	})
	messageValue = append(messageValue, map[string]any{
		"itemName": hamburger,
		"quantity": 2,
	})

	message := givenKafkaMessage(t, messageValue)

	kitchenService := stubs2.NewStubService()
	repositoryStub := stubs2.NewStubRepositoryWithOrder(givenExistingOrder())

	sut := &PackItemCommand{
		Stack:          s,
		Repository:     repositoryStub,
		KitchenService: kitchenService,
		Message:        message,
	}

	expectedPackedItems := []i.Item{
		{Name: cheeseburger, Quantity: 1},
		{Name: spicyStripes, Quantity: 2},
		{Name: mcSpicy, Quantity: 3},
		{Name: hamburger, Quantity: 2},
	}

	// when
	result, err := sut.Execute(context.Background())

	// then
	assert.True(t, result, "Successful result")
	assert.Nil(t, err, "No error message")

	// and
	assert.Equal(t, 2, repositoryStub.CalledCnt(), "Order Repository have been called once")

	upsertArgs := repositoryStub.GetUpsertArgs()
	assert.Len(t, upsertArgs, 1)

	actualOrder := upsertArgs[0]
	assert.Equal(t, actualOrder.Status, m.InProgress)
	assert.Equal(t, actualOrder.PackedItems, expectedPackedItems)

	// and
	assert.Equal(t, 1, s.GetCurrent(hamburger))
	assert.Equal(t, 2, s.GetCurrent(cheeseburger))
	assert.Equal(t, 2, s.GetCurrent(spicyStripes))
	assert.Equal(t, 1, s.GetCurrent(mcSpicy))

	// and
	assert.Equal(t, 0, kitchenService.CalledCnt(), "No items have not been requested")
}

func shouldFinishPackingOrderWhenLastItemsCameFromKitchen(t *testing.T) {
	// given
	s := stack.NewStack(stack.CleanStack())
	s.AddMany(hamburger, 3)
	s.AddMany(cheeseburger, 2)
	s.AddMany(mcSpicy, 4)
	s.AddMany(spicyStripes, 8)

	messageValue := make([]map[string]any, 0)
	messageValue = append(messageValue, map[string]any{
		"itemName": mcSpicy,
		"quantity": 3,
	})
	messageValue = append(messageValue, map[string]any{
		"itemName": hamburger,
		"quantity": 2,
	})
	messageValue = append(messageValue, map[string]any{
		"itemName": cheeseburger,
		"quantity": 2,
	})
	messageValue = append(messageValue, map[string]any{
		"itemName": spicyStripes,
		"quantity": 7,
	})

	message := givenKafkaMessage(t, messageValue)

	kitchenService := stubs2.NewStubService()
	repositoryStub := stubs2.NewStubRepositoryWithOrder(givenExistingOrder())

	sut := &PackItemCommand{
		Stack:          s,
		Repository:     repositoryStub,
		KitchenService: kitchenService,
		Message:        message,
	}

	expectedPackedItems := []i.Item{
		{Name: cheeseburger, Quantity: 1},
		{Name: spicyStripes, Quantity: 2},
		{Name: mcSpicy, Quantity: 3},
		{Name: hamburger, Quantity: 2},
		{Name: cheeseburger, Quantity: 1},
		{Name: spicyStripes, Quantity: 6},
	}

	// when
	result, err := sut.Execute(context.Background())

	// then
	assert.True(t, result, "Successful result")
	assert.Nil(t, err, "No error message")

	// and
	assert.Equal(t, 2, repositoryStub.CalledCnt(), "Order Repository have been called once")

	upsertArgs := repositoryStub.GetUpsertArgs()
	assert.Len(t, upsertArgs, 1)

	actualOrder := upsertArgs[0]
	assert.Equal(t, actualOrder.Status, m.Ready)
	assert.Equal(t, actualOrder.PackedItems, expectedPackedItems)

	// and
	assert.Equal(t, 1, s.GetCurrent(hamburger))
	assert.Equal(t, 1, s.GetCurrent(cheeseburger))
	assert.Equal(t, 2, s.GetCurrent(spicyStripes))
	assert.Equal(t, 1, s.GetCurrent(mcSpicy))

	// and
	assert.Equal(t, 0, kitchenService.CalledCnt(), "No items have not been requested")
}

func shouldRequestAdditionalItemWhenMoreAreNeeded(t *testing.T) {
	// given
	s := stack.NewStack(stack.CleanStack())
	s.AddMany(spicyStripes, 3)
	s.AddMany(cheeseburger, 2)

	messageValue := make([]map[string]any, 0)
	messageValue = append(messageValue, map[string]any{
		"itemName": spicyStripes,
		"quantity": 3,
	})

	message := givenKafkaMessage(t, messageValue)

	kitchenService := stubs2.NewStubService()
	repositoryStub := stubs2.NewStubRepositoryWithOrder(givenExistingOrder())

	sut := &PackItemCommand{
		Stack:          s,
		Repository:     repositoryStub,
		KitchenService: kitchenService,
		Message:        message,
	}

	expectedPackedItems := []i.Item{
		{Name: cheeseburger, Quantity: 1},
		{Name: spicyStripes, Quantity: 2},
		{Name: spicyStripes, Quantity: 3},
	}

	// when
	result, err := sut.Execute(context.Background())

	// then
	assert.True(t, result, "Successful result")
	assert.Nil(t, err, "No error message")

	// and
	assert.Equal(t, 2, repositoryStub.CalledCnt(), "Order Repository have been called once")

	upsertArgs := repositoryStub.GetUpsertArgs()
	assert.Len(t, upsertArgs, 1)

	actualOrder := upsertArgs[0]
	assert.Equal(t, actualOrder.Status, m.InProgress)
	assert.Equal(t, actualOrder.PackedItems, expectedPackedItems)

	// and
	assert.Equal(t, 2, s.GetCurrent(cheeseburger))
	assert.Equal(t, 0, s.GetCurrent(spicyStripes))

	// and
	assert.Equal(t, 1, kitchenService.CalledCnt(), "No items have not been requested")
	assert.True(t, kitchenService.HaveBeenCalledWith(spicyStripes, 3, expectedOrderNumber))
}

func shouldFailWhenMessageValueIsEmpty(t *testing.T) {
	// given
	s := stack.NewStack(stack.CleanStack())
	message := givenKafkaMessage(t, make([]map[string]any, 0))

	kitchenService := stubs2.NewStubService()
	repositoryStub := stubs2.NewStubRepositoryWithOrder(givenExistingOrder())

	sut := &PackItemCommand{
		Stack:          s,
		Repository:     repositoryStub,
		KitchenService: kitchenService,
		Message:        message,
	}

	// when
	result, err := sut.Execute(context.Background())

	// then
	assert.False(t, result)
	assert.Equal(t, "event message is nil or empty", err.Error())

	// and
	assert.Equal(t, 1, repositoryStub.CalledCnt(), "Order Repository have not been called")
	assert.Len(t, repositoryStub.GetUpsertArgs(), 0)

	// and
	assert.Equal(t, 0, kitchenService.CalledCnt(), "No items have not been requested")
}

func shouldFailWhenMessageKeyMissingOrderNumber(t *testing.T) {
	// given
	s := stack.NewStack(stack.CleanStack())
	b, err := json.Marshal(make([]map[string]any, 0))
	if err != nil {
		assert.Fail(t, "Could not marshal Kafka message", err)
	}

	message := kafka.Message{
		Topic: "item-added-to-stack",
		Key:   []byte("12131"),
		Value: b,
	}

	kitchenService := stubs2.NewStubService()
	repositoryStub := stubs2.NewStubRepositoryWithOrder(givenExistingOrder())

	sut := &PackItemCommand{
		Stack:          s,
		Repository:     repositoryStub,
		KitchenService: kitchenService,
		Message:        message,
	}

	// when
	result, err := sut.Execute(context.Background())

	// then
	assert.False(t, result)
	assert.Equal(t, "cannot extract prefix from message key - `12131`", err.Error())

	// and
	assert.Equal(t, 0, repositoryStub.CalledCnt(), "Order Repository have not been called")
	assert.Len(t, repositoryStub.GetUpsertArgs(), 0)

	// and
	assert.Equal(t, 0, kitchenService.CalledCnt(), "No items have not been requested")
}

func shouldFailWhenMessageKeyHasInvalidOrderNumber(t *testing.T) {
	// given
	s := stack.NewStack(stack.CleanStack())
	b, err := json.Marshal(make([]map[string]any, 0))
	if err != nil {
		assert.Fail(t, "Could not marshal Kafka message", err)
	}

	message := kafka.Message{
		Topic: "item-added-to-stack",
		Key:   []byte("order-`1212"),
		Value: b,
	}

	kitchenService := stubs2.NewStubService()
	repositoryStub := stubs2.NewStubRepositoryWithOrder(givenExistingOrder())

	sut := &PackItemCommand{
		Stack:          s,
		Repository:     repositoryStub,
		KitchenService: kitchenService,
		Message:        message,
	}

	// when
	result, err := sut.Execute(context.Background())

	// then
	assert.False(t, result)
	assert.Equal(t, "cannot parse order number '`1212' to int64. strconv.ParseInt: parsing \"`1212\": invalid syntax", err.Error())

	// and
	assert.Equal(t, 0, repositoryStub.CalledCnt())
	assert.Len(t, repositoryStub.GetUpsertArgs(), 0)

	// and
	assert.Equal(t, 0, kitchenService.CalledCnt(), "No items have not been requested")
}

func givenExistingOrder() *m.Order {
	return &m.Order{
		OrderNumber: expectedOrderNumber,
		CustomerId:  10,
		Status:      m.Requested,
		Items:       orderItems,
		PackedItems: []i.Item{
			{Name: cheeseburger, Quantity: 1},
			{Name: spicyStripes, Quantity: 2},
		},
	}
}

func givenKafkaMessage(t *testing.T, messageValue []map[string]any) kafka.Message {
	b, err := json.Marshal(messageValue)
	if err != nil {
		assert.Fail(t, "Could not marshal Kafka message", err)
	}

	message := kafka.Message{
		Topic: "item-added-to-stack",
		Key:   []byte(fmt.Sprintf("order-%d", expectedOrderNumber)),
		Value: b,
	}
	return message
}
