package command

import (
	"github.com/stretchr/testify/assert"
	i "mc-burger-orders/item"
	m "mc-burger-orders/order/model"
	"mc-burger-orders/stack"
	"testing"
)

type StubKitchenService struct {
	methodCalled []map[string]any
}

func (s *StubKitchenService) Request(itemName string, quantity int) {
	args := map[string]any{
		"itemName": itemName,
		"quantity": quantity,
	}
	s.methodCalled = append(s.methodCalled, args)
}

func (s *StubKitchenService) CalledCnt() int {
	return len(s.methodCalled)
}

func (s *StubKitchenService) HaveBeenCalledWith(itemName string, quantity int) bool {
	r := false

	for _, args := range s.methodCalled {
		if args["itemName"] == itemName && args["quantity"] == quantity {
			r = true
		}
	}

	return r
}

func Test_CreateNewOrder(t *testing.T) {
	// given
	repository := m.OrderRepository{}
	sk := stack.CleanStack()
	sk["hamburger"] = 3

	s := stack.NewStack(sk)
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

	stubKitchenService := &StubKitchenService{}
	command := &NewRequestCommand{
		Repository:     repository,
		Stack:          s,
		KitchenService: stubKitchenService,
		NewOrder:       newOrder,
	}

	// when
	order, err := command.Execute()

	// then
	assert.Nil(t, err)

	assert.Equal(t, 1010, order.ID)
	assert.Equal(t, 10, order.CustomerId)
	assert.Equal(t, "REQUESTED", order.Status.Value())

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
	repository := m.OrderRepository{}
	sk := stack.CleanStack()
	sk["hamburger"] = 1

	s := stack.NewStack(sk)
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

	stubKitchenService := &StubKitchenService{}
	command := &NewRequestCommand{
		Repository:     repository,
		Stack:          s,
		KitchenService: stubKitchenService,
		NewOrder:       newOrder,
	}

	// when
	order, err := command.Execute()

	// then
	assert.Nil(t, err)
	assert.Equal(t, 1010, order.ID)
	assert.Equal(t, 10, order.CustomerId)
	assert.Equal(t, "REQUESTED", order.Status.Value())

	assert.Equal(t, 2, len(order.Items))

	assert.Equal(t, 2, len(order.PackedItems))
	assert.Equal(t, expectedPackedItems, order.PackedItems)

	// and
	assert.Equal(t, 0, s.GetCurrent("hamburger"))

	// and
	assert.True(t, stubKitchenService.HaveBeenCalledWith("hamburger", 2), "Kitchen Service called with Hamburger requests")
}

func Test_DontPackItemsWhenNonIsInStack(t *testing.T) {
	// given
	repository := m.OrderRepository{}
	s := stack.NewStack(stack.CleanStack())
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
	stubKitchenService := &StubKitchenService{}

	command := &NewRequestCommand{
		Repository:     repository,
		Stack:          s,
		KitchenService: stubKitchenService,
		NewOrder:       newOrder,
	}

	// when
	order, err := command.Execute()

	// then
	assert.Nil(t, err)
	assert.Equal(t, 1010, order.ID)
	assert.Equal(t, 10, order.CustomerId)
	assert.Equal(t, "REQUESTED", order.Status.Value())

	assert.Equal(t, 2, len(order.Items))

	assert.Equal(t, 0, len(order.PackedItems))

	// and
	assert.Equal(t, 0, s.GetCurrent("hamburger"))

	// and
	assert.True(t, stubKitchenService.HaveBeenCalledWith("hamburger", 2), "Kitchen Service called with Hamburger requests")
	assert.True(t, stubKitchenService.HaveBeenCalledWith("fries", 1), "Kitchen Service called with Fries requests")
}
