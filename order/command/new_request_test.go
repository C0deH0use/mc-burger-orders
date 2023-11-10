package command

import (
	"context"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	i "mc-burger-orders/item"
	m "mc-burger-orders/order/model"
	"mc-burger-orders/stack"
	"testing"
)

type StubService struct {
	o            *m.Order
	err          error
	methodCalled []map[string]interface{}
}

func (s *StubService) Request(itemName string, quantity int) {
	args := map[string]interface{}{
		"itemName": itemName,
		"quantity": quantity,
	}
	s.methodCalled = append(s.methodCalled, args)
}

func (s *StubService) CalledCnt() int {
	return len(s.methodCalled)
}

func (s *StubService) HaveBeenCalledWith(itemName string, quantity int) bool {
	r := false

	for _, args := range s.methodCalled {
		if args["itemName"] == itemName && args["quantity"] == quantity {
			r = true
		}
	}

	return r
}

func (s *StubService) InsertOrUpdate(ctx context.Context, order m.Order) (*m.Order, error) {
	s.methodCalled = append(s.methodCalled, map[string]interface{}{"InsertOrUpdate": order})
	return s.o, s.err
}

func (s *StubService) FetchById(ctx context.Context, id interface{}) (*m.Order, error) {
	s.methodCalled = append(s.methodCalled, map[string]interface{}{"FetchById": id})
	return s.o, nil
}
func (s *StubService) FetchMany(ctx context.Context) ([]m.Order, error) {
	s.methodCalled = append(s.methodCalled, map[string]interface{}{"FetchMany": nil})
	return []m.Order{*s.o}, nil
}

func (s *StubService) GetOrderArgs() []m.Order {
	var u []m.Order

	for _, methodInvocation := range s.methodCalled {
		if v, exists := methodInvocation["InsertOrUpdate"]; exists {
			var o = m.Order{}
			dbByte, _ := json.Marshal(v)
			err := json.Unmarshal(dbByte, &o)
			if err != nil {
				panic(err)
			}
			u = append(u, o)
		}
	}

	return u
}

func Test_CreateNewOrder(t *testing.T) {
	// given
	sk := stack.CleanStack()
	sk["hamburger"] = 3

	s := stack.NewStack(sk)
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

	stubKitchenService := &StubService{}
	stubRepository := &StubService{o: &m.Order{
		OrderNumber: expectedOrderNumber,
		CustomerId:  10,
		Status:      m.Requested,
		Items:       newOrder.Items,
		PackedItems: []i.Item{},
	}}
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
	assert.Len(t, stubRepository.GetOrderArgs(), 1)
	order := stubRepository.GetOrderArgs()[0]

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
	sk := stack.CleanStack()
	sk["hamburger"] = 1

	s := stack.NewStack(sk)
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

	stubKitchenService := &StubService{}
	stubRepository := &StubService{o: &m.Order{
		OrderNumber: expectedOrderNumber,
		CustomerId:  10,
		Status:      m.Requested,
		Items:       newOrder.Items,
		PackedItems: []i.Item{},
	}}
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
	assert.Len(t, stubRepository.GetOrderArgs(), 1)
	order := stubRepository.GetOrderArgs()[0]

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
	assert.True(t, stubKitchenService.HaveBeenCalledWith("hamburger", 2), "Kitchen Service called with Hamburger requests")
}

func Test_DontPackItemsWhenNonIsInStack(t *testing.T) {
	// given
	s := stack.NewStack(stack.CleanStack())
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
	stubKitchenService := &StubService{}
	stubRepository := &StubService{o: &m.Order{
		OrderNumber: expectedOrderNumber,
		CustomerId:  10,
		Status:      m.Requested,
		Items:       newOrder.Items,
		PackedItems: []i.Item{},
	}}
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
	assert.Len(t, stubRepository.GetOrderArgs(), 1)
	order := stubRepository.GetOrderArgs()[0]

	// and
	assert.Equal(t, expectedOrderNumber, order.OrderNumber)
	assert.Equal(t, 10, order.CustomerId)
	assert.Equal(t, m.OrderStatus("REQUESTED"), order.Status)

	assert.Equal(t, 2, len(order.Items))

	assert.Equal(t, 0, len(order.PackedItems))

	// and
	assert.Equal(t, 0, s.GetCurrent("hamburger"))

	// and
	assert.True(t, stubKitchenService.HaveBeenCalledWith("hamburger", 2), "Kitchen Service called with Hamburger requests")
	assert.True(t, stubKitchenService.HaveBeenCalledWith("fries", 1), "Kitchen Service called with Fries requests")
}
