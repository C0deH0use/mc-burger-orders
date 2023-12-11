package order

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/segmentio/kafka-go"
	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"
	"mc-burger-orders/command"
	i "mc-burger-orders/kitchen/item"
	"mc-burger-orders/shelf"
	"testing"
	"time"
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
	t.Run("should add new items added to shelf when command executed", shouldPackItemPointedInMessage)
	t.Run("should set to READY when all items are packed", shouldFinishPackingOrderWhenLastItemsCameFromKitchen)
	t.Run("should try and pack items to multiple orders when new items are added to shelf", shouldPackMultipleOrdersWhenMultipleItemsAdded)
	t.Run("should pack items of other orders when first order already is packed with the item that was added to shelf", shouldPackOtherOrdersWhenTheFirstOneIsAlreadyPackedByItem)
	t.Run("should pack available items and request new when not all items are available", shouldRequestAdditionalItemWhenMoreAreNeeded)
	t.Run("should fail when message value is empty", shouldFailWhenMessageValueIsEmpty)
}

func shouldPackItemPointedInMessage(t *testing.T) {
	// given
	s := shelf.NewEmptyShelf()
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

	kitchenService := NewOrderService()
	statusEmitter := NewOrderService()
	repositoryStub := GivenRepository()
	repositoryStub.ReturnOrders(givenExistingOrder())

	sut := &PackItemCommand{
		Shelf:          s,
		Repository:     repositoryStub,
		KitchenService: kitchenService,
		StatusEmitter:  statusEmitter,
	}

	expectedPackedItems := []i.Item{
		{Name: cheeseburger, Quantity: 1},
		{Name: spicyStripes, Quantity: 2},
		{Name: mcSpicy, Quantity: 3},
		{Name: hamburger, Quantity: 2},
	}
	commandResults := make(chan command.TypedResult)

	// when
	go sut.Execute(context.Background(), message, commandResults)

	// then
	commandResult := <-commandResults
	assert.True(t, commandResult.Result)
	assert.Nil(t, commandResult.Error)

	// and
	assert.Equal(t, 4, repositoryStub.CalledCnt())

	upsertArgs := repositoryStub.GetUpsertArgs()
	assert.Len(t, upsertArgs, 2)

	assert.Equal(t, upsertArgs[0].Status, InProgress)
	assert.Len(t, upsertArgs[0].PackedItems, 3)

	assert.Equal(t, upsertArgs[1].Status, InProgress)
	assert.Len(t, upsertArgs[1].PackedItems, 4)
	assert.Equal(t, upsertArgs[1].PackedItems, expectedPackedItems)

	// and
	assert.Equal(t, 1, s.GetCurrent(hamburger))
	assert.Equal(t, 2, s.GetCurrent(cheeseburger))
	assert.Equal(t, 2, s.GetCurrent(spicyStripes))
	assert.Equal(t, 1, s.GetCurrent(mcSpicy))

	// and
	assert.Equal(t, 0, kitchenService.CalledCnt())

	// and
	assert.Equal(t, 1, statusEmitter.CalledCnt())
	assert.True(t, statusEmitter.HaveBeenCalledWith(StatusUpdateMatchingFnc(InProgress)))
	close(commandResults)
}

func shouldFinishPackingOrderWhenLastItemsCameFromKitchen(t *testing.T) {
	// given
	s := shelf.NewEmptyShelf()
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

	kitchenService := NewOrderService()
	statusEmitter := NewOrderService()
	repositoryStub := GivenRepository()
	repositoryStub.ReturnOrders(givenExistingOrder())

	sut := &PackItemCommand{
		Shelf:          s,
		Repository:     repositoryStub,
		KitchenService: kitchenService,
		StatusEmitter:  statusEmitter,
	}

	expectedPackedItems := []i.Item{
		{Name: cheeseburger, Quantity: 1},
		{Name: spicyStripes, Quantity: 2},
		{Name: mcSpicy, Quantity: 3},
		{Name: hamburger, Quantity: 2},
		{Name: cheeseburger, Quantity: 1},
		{Name: spicyStripes, Quantity: 6},
	}
	commandResults := make(chan command.TypedResult)

	// when
	go sut.Execute(context.Background(), message, commandResults)

	// then
	commandResult := <-commandResults
	assert.True(t, commandResult.Result)
	assert.Nil(t, commandResult.Error)

	// and
	assert.Equal(t, 8, repositoryStub.CalledCnt())

	upsertArgs := repositoryStub.GetUpsertArgs()
	assert.Len(t, upsertArgs, 4)

	assert.Equal(t, upsertArgs[0].Status, InProgress)
	assert.Len(t, upsertArgs[0].PackedItems, 3)

	assert.Equal(t, upsertArgs[1].Status, InProgress)
	assert.Len(t, upsertArgs[1].PackedItems, 4)

	assert.Equal(t, upsertArgs[2].Status, InProgress)
	assert.Len(t, upsertArgs[2].PackedItems, 5)

	assert.Equal(t, upsertArgs[3].Status, Ready)
	assert.Equal(t, upsertArgs[3].PackedItems, expectedPackedItems)

	// and
	assert.Equal(t, 1, s.GetCurrent(hamburger))
	assert.Equal(t, 1, s.GetCurrent(cheeseburger))
	assert.Equal(t, 2, s.GetCurrent(spicyStripes))
	assert.Equal(t, 1, s.GetCurrent(mcSpicy))

	// and
	assert.Equal(t, 0, kitchenService.CalledCnt(), "No items have not been requested")

	// and
	assert.Equal(t, 2, statusEmitter.CalledCnt())
	assert.True(t, statusEmitter.HaveBeenCalledWith(StatusUpdateMatchingFnc(InProgress)))
	assert.True(t, statusEmitter.HaveBeenCalledWith(StatusUpdateMatchingFnc(Ready)))
	close(commandResults)
}

func shouldPackMultipleOrdersWhenMultipleItemsAdded(t *testing.T) {
	// given
	s := shelf.NewEmptyShelf()
	s.AddMany(spicyStripes, 18)
	s.AddMany(cheeseburger, 8)
	s.AddMany(hamburger, 8)

	messageValue := make([]map[string]any, 0)
	messageValue = append(messageValue, map[string]any{
		"itemName": spicyStripes,
		"quantity": 18,
	})
	messageValue = append(messageValue, map[string]any{
		"itemName": cheeseburger,
		"quantity": 8,
	})
	messageValue = append(messageValue, map[string]any{
		"itemName": hamburger,
		"quantity": 8,
	})
	message := givenKafkaMessage(t, messageValue)

	secondOrderNumber := int64(1012)
	thirdOrderNumber := int64(1013)
	existingOrders := []*Order{
		givenExistingOrder(),
		{
			OrderNumber: secondOrderNumber,
			CustomerId:  10,
			Status:      Requested,
			Items: []i.Item{
				{Name: hamburger, Quantity: 2},
				{Name: cheeseburger, Quantity: 2},
			},
		},
		{
			OrderNumber: thirdOrderNumber,
			CustomerId:  10,
			Status:      InProgress,
			Items: []i.Item{
				{Name: hamburger, Quantity: 2},
				{Name: cheeseburger, Quantity: 2},
				{Name: spicyStripes, Quantity: 8},
			},
			PackedItems: []i.Item{
				{Name: cheeseburger, Quantity: 2},
			},
		},
	}

	kitchenService := NewOrderService()
	statusEmitter := NewOrderService()
	repositoryStub := GivenRepository()
	repositoryStub.ReturnOrders(existingOrders...)

	sut := &PackItemCommand{
		Shelf:          s,
		Repository:     repositoryStub,
		KitchenService: kitchenService,
		StatusEmitter:  statusEmitter,
	}
	commandResults := make(chan command.TypedResult)

	// when
	go sut.Execute(context.Background(), message, commandResults)

	// then
	commandResult := <-commandResults
	assert.True(t, commandResult.Result)
	assert.Nil(t, commandResult.Error)

	// and
	upsertArgs := repositoryStub.GetUpsertArgs()
	assert.Len(t, upsertArgs, 8)

	firstOrder := getLastOrder(expectedOrderNumber, upsertArgs...)
	assert.Equal(t, firstOrder.Status, InProgress)
	assert.Equal(t, 0, getMissingItemsCount(firstOrder, hamburger))
	assert.Equal(t, 0, getMissingItemsCount(firstOrder, cheeseburger))
	assert.Equal(t, 3, getMissingItemsCount(firstOrder, mcSpicy))
	assert.Equal(t, 0, getMissingItemsCount(firstOrder, spicyStripes))

	secondOrder := getLastOrder(secondOrderNumber, upsertArgs...)
	assert.Equal(t, secondOrder.Status, Ready)
	assert.Equal(t, 0, getMissingItemsCount(secondOrder, hamburger))
	assert.Equal(t, 0, getMissingItemsCount(secondOrder, cheeseburger))

	thirdOrder := getLastOrder(thirdOrderNumber, upsertArgs...)
	assert.Equal(t, thirdOrder.Status, Ready)
	assert.Equal(t, 0, getMissingItemsCount(thirdOrder, hamburger))
	assert.Equal(t, 0, getMissingItemsCount(thirdOrder, cheeseburger))
	assert.Equal(t, 0, getMissingItemsCount(thirdOrder, spicyStripes))

	// and
	assert.Equal(t, 5, s.GetCurrent(cheeseburger))
	assert.Equal(t, 2, s.GetCurrent(hamburger))
	assert.Equal(t, 4, s.GetCurrent(spicyStripes))
	assert.Equal(t, 0, s.GetCurrent(mcSpicy))

	// and
	assert.Equal(t, 0, kitchenService.CalledCnt())

	// and
	assert.Equal(t, 4, statusEmitter.CalledCnt())
	assert.True(t, statusEmitter.HaveBeenCalledWith(StatusUpdateMatchingFnc(InProgress)))
	assert.True(t, statusEmitter.HaveBeenCalledWith(StatusUpdateMatchingFnc(Ready)))
	close(commandResults)
}

func shouldPackOtherOrdersWhenTheFirstOneIsAlreadyPackedByItem(t *testing.T) {
	// given
	s := shelf.NewEmptyShelf()
	s.AddMany(spicyStripes, 10)

	messageValue := make([]map[string]any, 0)
	messageValue = append(messageValue, map[string]any{
		"itemName": spicyStripes,
		"quantity": 8,
	})
	message := givenKafkaMessage(t, messageValue)

	secondOrderNumber := int64(1012)
	thirdOrderNumber := int64(1013)
	existingOrders := []*Order{
		{
			OrderNumber: expectedOrderNumber,
			CustomerId:  10,
			Status:      Requested,
			Items:       orderItems,
			PackedItems: []i.Item{
				{Name: cheeseburger, Quantity: 1},
				{Name: spicyStripes, Quantity: 8},
			},
		},
		{
			OrderNumber: secondOrderNumber,
			CustomerId:  10,
			Status:      Requested,
			Items: []i.Item{
				{Name: hamburger, Quantity: 2},
				{Name: spicyStripes, Quantity: 8},
			},
		},
		{
			OrderNumber: thirdOrderNumber,
			CustomerId:  10,
			Status:      InProgress,
			Items: []i.Item{
				{Name: hamburger, Quantity: 2},
				{Name: cheeseburger, Quantity: 2},
				{Name: spicyStripes, Quantity: 8},
			},
			PackedItems: []i.Item{
				{Name: cheeseburger, Quantity: 2},
			},
		},
	}

	kitchenService := NewOrderService()
	statusEmitter := NewOrderService()
	repositoryStub := GivenRepository()
	repositoryStub.ReturnOrders(existingOrders...)

	sut := &PackItemCommand{
		Shelf:          s,
		Repository:     repositoryStub,
		KitchenService: kitchenService,
		StatusEmitter:  statusEmitter,
	}
	commandResults := make(chan command.TypedResult)

	// when
	go sut.Execute(context.Background(), message, commandResults)

	// then
	commandResult := <-commandResults
	assert.True(t, commandResult.Result)
	assert.Nil(t, commandResult.Error)

	// and
	upsertArgs := repositoryStub.GetUpsertArgs()
	assert.Len(t, upsertArgs, 3)

	firstOrder := getLastOrder(expectedOrderNumber, upsertArgs...)
	assert.Equal(t, firstOrder.Status, InProgress)
	assert.Equal(t, 0, getMissingItemsCount(firstOrder, spicyStripes))

	secondOrder := getLastOrder(secondOrderNumber, upsertArgs...)
	assert.Equal(t, secondOrder.Status, InProgress)
	assert.Equal(t, 0, getMissingItemsCount(secondOrder, spicyStripes))

	thirdOrder := getLastOrder(thirdOrderNumber, upsertArgs...)
	assert.Equal(t, thirdOrder.Status, InProgress)
	assert.Equal(t, 6, getMissingItemsCount(thirdOrder, spicyStripes))

	// and
	assert.Equal(t, 0, s.GetCurrent(spicyStripes))

	// and
	assert.Equal(t, 1, kitchenService.CalledCnt())
	assert.True(t, kitchenService.HaveBeenCalledWith(RequestMatchingFnc(spicyStripes, 6)))

	// and
	assert.Equal(t, 2, statusEmitter.CalledCnt())
	assert.True(t, statusEmitter.HaveBeenCalledWith(StatusUpdateMatchingFnc(InProgress)))
	close(commandResults)
}

func shouldRequestAdditionalItemWhenMoreAreNeeded(t *testing.T) {
	// given
	s := shelf.NewEmptyShelf()
	s.AddMany(spicyStripes, 3)
	s.AddMany(cheeseburger, 2)

	messageValue := make([]map[string]any, 0)
	messageValue = append(messageValue, map[string]any{
		"itemName": spicyStripes,
		"quantity": 3,
	})

	message := givenKafkaMessage(t, messageValue)

	kitchenService := NewOrderService()
	statusEmitter := NewOrderService()
	repositoryStub := GivenRepository()
	repositoryStub.ReturnOrders(givenExistingOrder())

	sut := &PackItemCommand{
		Shelf:          s,
		Repository:     repositoryStub,
		KitchenService: kitchenService,
		StatusEmitter:  statusEmitter,
	}

	expectedPackedItems := []i.Item{
		{Name: cheeseburger, Quantity: 1},
		{Name: spicyStripes, Quantity: 2},
		{Name: spicyStripes, Quantity: 3},
	}
	commandResults := make(chan command.TypedResult)

	// when
	go sut.Execute(context.Background(), message, commandResults)

	// then
	commandResult := <-commandResults
	assert.True(t, commandResult.Result)
	assert.Nil(t, commandResult.Error)

	// and
	assert.Equal(t, 2, repositoryStub.CalledCnt())

	upsertArgs := repositoryStub.GetUpsertArgs()
	assert.Len(t, upsertArgs, 1)

	actualOrder := upsertArgs[0]
	assert.Equal(t, actualOrder.Status, InProgress)
	assert.Equal(t, actualOrder.PackedItems, expectedPackedItems)

	// and
	assert.Equal(t, 2, s.GetCurrent(cheeseburger))
	assert.Equal(t, 0, s.GetCurrent(spicyStripes))

	// and
	assert.Equal(t, 1, kitchenService.CalledCnt())
	assert.True(t, kitchenService.HaveBeenCalledWith(RequestMatchingFnc(spicyStripes, 3)))

	// and
	assert.Equal(t, 1, statusEmitter.CalledCnt())
	assert.True(t, statusEmitter.HaveBeenCalledWith(StatusUpdateMatchingFnc(InProgress)))
	close(commandResults)
}

func shouldFailWhenMessageValueIsEmpty(t *testing.T) {
	// given
	s := shelf.NewEmptyShelf()
	message := givenKafkaMessage(t, make([]map[string]any, 0))

	kitchenService := NewOrderService()
	statusEmitter := NewOrderService()
	repositoryStub := GivenRepository()
	repositoryStub.ReturnOrders(givenExistingOrder())

	sut := &PackItemCommand{
		Shelf:          s,
		Repository:     repositoryStub,
		StatusEmitter:  statusEmitter,
		KitchenService: kitchenService,
	}
	commandResults := make(chan command.TypedResult)

	// when
	go sut.Execute(context.Background(), message, commandResults)

	// then
	commandResult := <-commandResults
	assert.False(t, commandResult.Result)
	assert.Equal(t, "event message is nil or empty", commandResult.Error.Error())

	// and
	assert.Equal(t, 0, repositoryStub.CalledCnt(), "Order Repository have not been called")
	assert.Len(t, repositoryStub.GetUpsertArgs(), 0)

	// and
	assert.Equal(t, 0, kitchenService.CalledCnt(), "No items have not been requested")

	// and
	assert.Equal(t, 0, statusEmitter.CalledCnt())
	close(commandResults)
}

func givenExistingOrder() *Order {
	return &Order{
		OrderNumber: expectedOrderNumber,
		CustomerId:  10,
		Status:      Requested,
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
	headers := make([]kafka.Header, 0)
	headers = append(headers, kafka.Header{Key: "order", Value: []byte(cast.ToString(expectedOrderNumber))})
	headers = append(headers, kafka.Header{Key: "event", Value: []byte("item-added-to-shelf")})

	message := kafka.Message{
		Headers: headers,
		Topic:   "some-kafka-topic",
		Key:     []byte(fmt.Sprintf("%d", time.Now().UnixNano())),
		Value:   b,
	}
	return message
}

func getLastOrder(orderNumber int64, orders ...Order) Order {
	ordersPerNumber := make([]Order, 0)

	for _, o := range orders {
		if o.OrderNumber == orderNumber {
			ordersPerNumber = append(ordersPerNumber, o)
		}
	}

	return ordersPerNumber[len(ordersPerNumber)-1]
}

func getMissingItemsCount(order Order, itemName string) int {
	if val, err := order.GetMissingItemsCount(itemName); err == nil {
		return val
	}
	return -1
}
