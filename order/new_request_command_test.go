package order

import (
	"context"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	cmd "mc-burger-orders/command"
	i "mc-burger-orders/kitchen/item"
	"mc-burger-orders/shelf"
	"sync"
	"testing"
)

func Test_CreateNewOrder(t *testing.T) {
	// given
	s := shelf.NewEmptyShelf()
	s.AddMany("hamburger", 3)

	expectedOrderNumber := int64(1010)

	newOrder := NewOrder{
		CustomerId: 10,
		Items: []i.Item{
			{
				Name:     "hamburger",
				Quantity: 2,
			},
			{
				Name:     "ice-cream",
				Quantity: 2,
			},
		},
	}
	expectedPackedItems := []i.Item{
		{
			Name:     "hamburger",
			Quantity: 2,
		},
		{
			Name:     "ice-cream",
			Quantity: 2,
		},
	}

	stubRepository := GivenRepository()
	stubKitchenService := NewStubService()

	statusUpdateWg := &sync.WaitGroup{}
	statusUpdateWg.Add(2)
	stubStatusEmitter := NewStubService()
	stubStatusEmitter.WithWaitGroup(statusUpdateWg)

	command := &NewRequestCommand{
		Repository:     stubRepository,
		Shelf:          s,
		KitchenService: stubKitchenService,
		StatusEmitter:  stubStatusEmitter,
		OrderNumber:    expectedOrderNumber,
		NewOrder:       newOrder,
	}
	commandResults := make(chan cmd.TypedResult)

	// when
	go command.Execute(context.Background(), kafka.Message{}, commandResults)

	// then
	commandResult := <-commandResults
	assert.True(t, commandResult.Result)
	assert.Nil(t, commandResult.Error)

	// and
	assert.Len(t, stubRepository.GetUpsertArgs(), 2)
	updateOrderArg := stubRepository.GetUpsertArgs()[1]

	assert.Equal(t, expectedOrderNumber, updateOrderArg.OrderNumber)
	assert.Equal(t, 10, updateOrderArg.CustomerId)
	assert.Equal(t, OrderStatus("READY"), updateOrderArg.Status)

	assert.Equal(t, 2, len(updateOrderArg.Items))

	assert.Equal(t, 2, len(updateOrderArg.PackedItems))
	assert.Equal(t, expectedPackedItems, updateOrderArg.PackedItems)

	// and
	assert.Equal(t, 1, s.GetCurrent("hamburger"))

	// and
	assert.Empty(t, stubKitchenService.CalledCnt())

	// and
	statusUpdateWg.Wait()

	t.Log("StatusUpdatedEvent Args:", stubStatusEmitter.MethodCalled)
	assert.Len(t, stubStatusEmitter.GetStatusUpdatedEventArgs(), 2)
	assert.True(t, stubStatusEmitter.HaveBeenCalledWith(StatusUpdateMatchingFnc(Requested)))
	assert.True(t, stubStatusEmitter.HaveBeenCalledWith(StatusUpdateMatchingFnc(Ready)))
	close(commandResults)
}

func Test_CreateNewOrderAndPackOnlyTheseItemsThatAreAvailable(t *testing.T) {
	// given
	s := shelf.NewEmptyShelf()
	s.Add("hamburger")

	expectedOrderNumber := int64(1010)

	newOrder := NewOrder{
		CustomerId: 10,
		Items: []i.Item{
			{
				Name:     "hamburger",
				Quantity: 2,
			},
			{
				Name:     "ice-cream",
				Quantity: 2,
			},
		},
	}
	expectedPackedItems := []i.Item{
		{
			Name:     "hamburger",
			Quantity: 1,
		},
		{
			Name:     "ice-cream",
			Quantity: 2,
		},
	}

	kitchenWg := &sync.WaitGroup{}
	kitchenWg.Add(1)
	stubKitchenService := NewStubService()
	stubKitchenService.WithWaitGroup(kitchenWg)

	statusUpdateWg := &sync.WaitGroup{}
	statusUpdateWg.Add(2)
	stubStatusEmitter := NewStubService()
	stubStatusEmitter.WithWaitGroup(statusUpdateWg)

	expectedOrder := &Order{
		OrderNumber: expectedOrderNumber,
		CustomerId:  10,
		Status:      Requested,
		Items:       newOrder.Items,
		PackedItems: []i.Item{},
	}
	stubRepository := GivenRepository()
	stubRepository.ReturnOrders(expectedOrder)

	command := &NewRequestCommand{
		Repository:     stubRepository,
		Shelf:          s,
		KitchenService: stubKitchenService,
		StatusEmitter:  stubStatusEmitter,
		OrderNumber:    expectedOrderNumber,
		NewOrder:       newOrder,
	}
	commandResults := make(chan cmd.TypedResult)

	// when
	go command.Execute(context.Background(), kafka.Message{}, commandResults)

	// then
	commandResult := <-commandResults

	assert.True(t, commandResult.Result)
	assert.Nil(t, commandResult.Error)

	// and
	assert.Len(t, stubRepository.GetUpsertArgs(), 2)
	order := stubRepository.GetUpsertArgs()[1]

	// and
	assert.Equal(t, expectedOrderNumber, order.OrderNumber)
	assert.Equal(t, 10, order.CustomerId)
	assert.Equal(t, OrderStatus("IN_PROGRESS"), order.Status)

	assert.Equal(t, 2, len(order.Items))

	assert.Equal(t, 2, len(order.PackedItems))
	assert.Equal(t, expectedPackedItems, order.PackedItems)

	// and
	assert.Equal(t, 0, s.GetCurrent("hamburger"))

	// and
	kitchenWg.Wait()
	assert.True(t, stubKitchenService.HaveBeenCalledWith(RequestMatchingFnc("hamburger", 1)))

	// and
	statusUpdateWg.Wait()
	t.Log("StatusUpdatedEvent Args:", stubStatusEmitter.MethodCalled)

	assert.Len(t, stubStatusEmitter.GetStatusUpdatedEventArgs(), 2)
	assert.True(t, stubStatusEmitter.HaveBeenCalledWith(StatusUpdateMatchingFnc(Requested)))
	assert.True(t, stubStatusEmitter.HaveBeenCalledWith(StatusUpdateMatchingFnc(InProgress)))
	close(commandResults)
}

func Test_DontPackItemsWhenNonIsInStack(t *testing.T) {
	// given
	s := shelf.NewEmptyShelf()
	expectedOrderNumber := int64(1010)
	newOrder := NewOrder{
		CustomerId: 10,
		Items: []i.Item{
			{
				Name:     "hamburger",
				Quantity: 2,
			},
			{
				Name:     "fries",
				Quantity: 1,
			},
		},
	}
	stubRepository := GivenRepository()

	kitchenWg := &sync.WaitGroup{}
	kitchenWg.Add(2)
	stubKitchenService := NewStubService()
	stubKitchenService.WithWaitGroup(kitchenWg)

	statusUpdateWg := &sync.WaitGroup{}
	statusUpdateWg.Add(1)
	stubStatusEmitter := NewStubService()
	stubStatusEmitter.WithWaitGroup(statusUpdateWg)

	command := &NewRequestCommand{
		Repository:     stubRepository,
		Shelf:          s,
		KitchenService: stubKitchenService,
		StatusEmitter:  stubStatusEmitter,
		OrderNumber:    expectedOrderNumber,
		NewOrder:       newOrder,
	}
	commandResults := make(chan cmd.TypedResult)

	// when
	go command.Execute(context.Background(), kafka.Message{}, commandResults)

	// then
	commandResult := <-commandResults
	assert.True(t, commandResult.Result)
	assert.Nil(t, commandResult.Error)

	// and
	assert.Len(t, stubRepository.GetUpsertArgs(), 2)
	order := stubRepository.GetUpsertArgs()[1]

	// and
	assert.Equal(t, expectedOrderNumber, order.OrderNumber)
	assert.Equal(t, 10, order.CustomerId)
	assert.Equal(t, OrderStatus("REQUESTED"), order.Status)

	assert.Equal(t, 2, len(order.Items))

	assert.Equal(t, 0, len(order.PackedItems))

	// and
	assert.Equal(t, 0, s.GetCurrent("hamburger"))

	// and
	kitchenWg.Wait()
	assert.True(t, stubKitchenService.HaveBeenCalledWith(RequestMatchingFnc("hamburger", 2)))
	assert.True(t, stubKitchenService.HaveBeenCalledWith(RequestMatchingFnc("fries", 1)))

	// and
	statusUpdateWg.Wait()
	assert.Len(t, stubStatusEmitter.GetStatusUpdatedEventArgs(), 1)
	assert.True(t, stubStatusEmitter.HaveBeenCalledWith(StatusUpdateMatchingFnc(Requested)))
	close(commandResults)
}
