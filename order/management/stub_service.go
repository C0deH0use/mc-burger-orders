package management

import (
	"context"
	"log"
	"mc-burger-orders/order"
	"mc-burger-orders/testing/stubs"
	"sync"
)

type StubService struct {
	Wg *sync.WaitGroup
	stubs.DefaultStubService
}

type OrderQueryStubService struct {
	FindOrders []order.Order
	StubService
}

func NewKitchenStubService(waitG *sync.WaitGroup) *StubService {
	return &StubService{Wg: waitG, DefaultStubService: stubs.DefaultStubService{MethodCalled: make([]map[string]any, 0)}}
}

func NewStubQueryService() *OrderQueryStubService {
	return &OrderQueryStubService{FindOrders: make([]order.Order, 0), StubService: StubService{Wg: nil, DefaultStubService: stubs.DefaultStubService{MethodCalled: make([]map[string]any, 0)}}}
}

func (s *OrderQueryStubService) ReturnOnFindPackingOrders(orders []order.Order) {
	s.FindOrders = orders
}

func (s *StubService) RequestNew(ctx context.Context, itemName string, quantity int) error {
	args := map[string]any{
		"RequestNew": map[string]any{
			"itemName": itemName,
			"quantity": quantity,
		},
	}
	s.MethodCalled = append(s.MethodCalled, args)

	if s.Wg != nil {
		s.Wg.Done()
	}
	return nil
}

func (s *OrderQueryStubService) FetchOrdersForPacking(ctx context.Context) ([]order.Order, error) {
	args := map[string]any{
		"FetchOrdersForPacking": 1,
	}
	s.MethodCalled = append(s.MethodCalled, args)

	if s.Wg != nil {
		s.Wg.Done()
	}
	return s.FindOrders, nil
}

func RequestMatchingFnc(itemName string, quantity int) func(args map[string]any) bool {
	return func(args map[string]any) bool {
		value, exists := args["RequestNew"]
		if exists {
			innerVal := value.(map[string]any)
			argName := innerVal["itemName"]
			argQuantity := innerVal["quantity"]
			log.Printf("StubService methodCalled. %+v", args)
			b := argName == itemName && argQuantity == quantity
			return b
		}
		return false
	}
}
