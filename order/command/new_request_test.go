package command

import (
	"context"
	"github.com/stretchr/testify/assert"
	i "mc-burger-orders/item"
	m "mc-burger-orders/order/model"
	"mc-burger-orders/stack"
	stubs2 "mc-burger-orders/testing/stubs"
	"testing"
)

func Test_CreateNewOrder(t *testing.T) {
	// given
	s := stack.NewEmptyStack()
	s.AddMany("hamburger", 3)

	expectedOrderNumber := int64(1010)

	newOrder := m.NewOrder{
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

	stubKitchenService := stubs2.NewStubService()
	expectedOrder := &m.Order{
		OrderNumber: expectedOrderNumber,
		CustomerId:  10,
		Status:      m.Requested,
		Items:       newOrder.Items,
		PackedItems: []i.Item{},
	}
	stubRepository := stubs2.NewStubRepositoryWithOrder(expectedOrder)
	command := &NewRequestCommand{
		Repository:     stubRepository,
		Stack:          s,
		KitchenService: stubKitchenService,
		OrderNumber:    expectedOrderNumber,
		NewOrder:       newOrder,
	}

	// when
	result, err := command.Execute(context.TODO())

	// then
	assert.Nil(t, err)
	assert.True(t, result)

	// and
	assert.Len(t, stubRepository.GetUpsertArgs(), 1)
	order := stubRepository.GetUpsertArgs()[0]

	assert.Equal(t, expectedOrderNumber, order.OrderNumber)
	assert.Equal(t, 10, order.CustomerId)
	assert.Equal(t, m.OrderStatus("READY"), order.Status)

	assert.Equal(t, 2, len(order.Items))

	assert.Equal(t, 2, len(order.PackedItems))
	assert.Equal(t, expectedPackedItems, order.PackedItems)

	// and
	assert.Equal(t, 1, s.GetCurrent("hamburger"))

	// and
	assert.Empty(t, stubKitchenService.CalledCnt(), "Kitchen Service have not been called")
}

func Test_CreateNewOrderAndPackOnlyTheseItemsThatAreAvailable(t *testing.T) {
	// given
	s := stack.NewEmptyStack()
	s.Add("hamburger")

	expectedOrderNumber := int64(1010)

	newOrder := m.NewOrder{
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

	stubKitchenService := stubs2.NewStubService()
	expectedOrder := &m.Order{
		OrderNumber: expectedOrderNumber,
		CustomerId:  10,
		Status:      m.Requested,
		Items:       newOrder.Items,
		PackedItems: []i.Item{},
	}
	stubRepository := stubs2.NewStubRepositoryWithOrder(expectedOrder)
	command := &NewRequestCommand{
		Repository:     stubRepository,
		Stack:          s,
		KitchenService: stubKitchenService,
		OrderNumber:    expectedOrderNumber,
		NewOrder:       newOrder,
	}

	// when
	result, err := command.Execute(context.TODO())

	// then
	assert.Nil(t, err)
	assert.True(t, result)

	// and
	assert.Len(t, stubRepository.GetUpsertArgs(), 1)
	order := stubRepository.GetUpsertArgs()[0]

	// and
	assert.Equal(t, expectedOrderNumber, order.OrderNumber)
	assert.Equal(t, 10, order.CustomerId)
	assert.Equal(t, m.OrderStatus("IN_PROGRESS"), order.Status)

	assert.Equal(t, 2, len(order.Items))

	assert.Equal(t, 2, len(order.PackedItems))
	assert.Equal(t, expectedPackedItems, order.PackedItems)

	// and
	assert.Equal(t, 0, s.GetCurrent("hamburger"))

	// and
	assert.True(t, stubKitchenService.HaveBeenCalledWith(stubs2.RequestMatchingFnc("hamburger", 1, expectedOrderNumber)), "Kitchen Service called with Hamburger requests")
}

func Test_DontPackItemsWhenNonIsInStack(t *testing.T) {
	// given
	s := stack.NewEmptyStack()
	expectedOrderNumber := int64(1010)
	newOrder := m.NewOrder{
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
	stubKitchenService := stubs2.NewStubService()
	expectedOrder := &m.Order{
		OrderNumber: expectedOrderNumber,
		CustomerId:  10,
		Status:      m.Requested,
		Items:       newOrder.Items,
		PackedItems: []i.Item{},
	}
	stubRepository := stubs2.NewStubRepositoryWithOrder(expectedOrder)
	command := &NewRequestCommand{
		Repository:     stubRepository,
		Stack:          s,
		KitchenService: stubKitchenService,
		OrderNumber:    expectedOrderNumber,
		NewOrder:       newOrder,
	}

	// when
	result, err := command.Execute(context.TODO())

	// then
	assert.Nil(t, err)
	assert.True(t, result)

	// and
	assert.Len(t, stubRepository.GetUpsertArgs(), 1)
	order := stubRepository.GetUpsertArgs()[0]

	// and
	assert.Equal(t, expectedOrderNumber, order.OrderNumber)
	assert.Equal(t, 10, order.CustomerId)
	assert.Equal(t, m.OrderStatus("REQUESTED"), order.Status)

	assert.Equal(t, 2, len(order.Items))

	assert.Equal(t, 0, len(order.PackedItems))

	// and
	assert.Equal(t, 0, s.GetCurrent("hamburger"))

	// and
	assert.True(t, stubKitchenService.HaveBeenCalledWith(stubs2.RequestMatchingFnc("hamburger", 2, expectedOrderNumber)), "Kitchen Service called with Hamburger requests")
	assert.True(t, stubKitchenService.HaveBeenCalledWith(stubs2.RequestMatchingFnc("fries", 1, expectedOrderNumber)), "Kitchen Service called with Fries requests")
}
