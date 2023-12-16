package management

import (
	"context"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"mc-burger-orders/command"
	"mc-burger-orders/kitchen/item"
	"mc-burger-orders/order"
	"testing"
)

func TestCheckMissingItemsOnOrdersCommand_Execute(t *testing.T) {
	t.Run("should request new item when order is not ready yet", shouldRequestNewItemWhenOrderIsNotReadyYet)
	t.Run("should request multiple items when multiple orders arte not yet ready", shouldRequestMultipleItemsWhenMultipleOrdersArteNotYetReady)
	t.Run("should not request any items when no orders are missing any items", shouldNotRequestAnyItemsWhenNoOrdersReturnedByQueryService)
}

func shouldRequestNewItemWhenOrderIsNotReadyYet(t *testing.T) {
	// given
	stubQueryService := NewStubQueryService()
	kitchenStubService := NewKitchenStubService(nil)
	cmd := CheckMissingItemsOnOrdersCommand{
		queryService:   stubQueryService,
		kitchenService: kitchenStubService,
	}

	stubQueryService.ReturnOnFindPackingOrders([]order.Order{
		{OrderNumber: 1000, Items: []item.Item{{Name: "hamburger", Quantity: 1}, {Name: "fries", Quantity: 1}}, PackedItems: []item.Item{{Name: "fries", Quantity: 1}}},
	})
	commandResults := make(chan command.TypedResult)

	// when
	go cmd.Execute(context.Background(), kafka.Message{}, commandResults)

	// then
	result := <-commandResults

	assert.True(t, result.Result)
	assert.Equal(t, 1, stubQueryService.CalledCnt())

	assert.Equal(t, 1, kitchenStubService.CalledCnt())
	assert.True(t, kitchenStubService.HaveBeenCalledWith(RequestMatchingFnc("hamburger", 1)))
}

func shouldRequestMultipleItemsWhenMultipleOrdersArteNotYetReady(t *testing.T) {
	// given
	stubQueryService := NewStubQueryService()
	kitchenStubService := NewKitchenStubService(nil)
	cmd := CheckMissingItemsOnOrdersCommand{
		queryService:   stubQueryService,
		kitchenService: kitchenStubService,
	}

	stubQueryService.ReturnOnFindPackingOrders([]order.Order{
		{OrderNumber: 1000, Items: []item.Item{{Name: "hamburger", Quantity: 1}, {Name: "fries", Quantity: 1}}, PackedItems: []item.Item{{Name: "fries", Quantity: 1}}},
		{OrderNumber: 1001, Items: []item.Item{{Name: "cheeseburger", Quantity: 3}, {Name: "fries", Quantity: 1}, {Name: "coke", Quantity: 1}}, PackedItems: []item.Item{{Name: "cheeseburger", Quantity: 1}, {Name: "fries", Quantity: 1}, {Name: "fries", Quantity: 1}, {Name: "coke", Quantity: 1}}},
	})
	commandResults := make(chan command.TypedResult)

	// when
	go cmd.Execute(context.Background(), kafka.Message{}, commandResults)

	// then
	result := <-commandResults

	assert.True(t, result.Result)
	assert.Equal(t, 1, stubQueryService.CalledCnt())

	assert.Equal(t, 2, kitchenStubService.CalledCnt())
	assert.True(t, kitchenStubService.HaveBeenCalledWith(RequestMatchingFnc("hamburger", 1)))
	assert.True(t, kitchenStubService.HaveBeenCalledWith(RequestMatchingFnc("cheeseburger", 2)))
}

func shouldNotRequestAnyItemsWhenNoOrdersReturnedByQueryService(t *testing.T) {
	// given
	stubQueryService := NewStubQueryService()
	kitchenStubService := NewKitchenStubService(nil)
	cmd := CheckMissingItemsOnOrdersCommand{
		queryService:   stubQueryService,
		kitchenService: kitchenStubService,
	}

	commandResults := make(chan command.TypedResult)

	// when
	go cmd.Execute(context.Background(), kafka.Message{}, commandResults)

	// then
	result := <-commandResults

	assert.True(t, result.Result)
	assert.Equal(t, 1, stubQueryService.CalledCnt())

	assert.Equal(t, 0, kitchenStubService.CalledCnt())
}
