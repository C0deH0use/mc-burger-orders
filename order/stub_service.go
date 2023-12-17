package order

import (
	"context"
	"log"
	"mc-burger-orders/testing/stubs"
)

type StubService struct {
	stubs.DefaultStubService
}

func NewStubService() *StubService {
	return &StubService{stubs.DefaultStubService{MethodCalled: make([]map[string]any, 0)}}
}

func RequestMatchingFnc(itemName string, quantity int) func(args map[string]any) bool {
	return func(args map[string]any) bool {
		argName := args["itemName"]
		argQuantity := args["quantity"]
		log.Printf("StubService methodCalled. %+v", args)
		b := argName == itemName && argQuantity == quantity
		return b
	}
}

func StatusUpdateMatchingFnc(status OrderStatus) func(args map[string]any) bool {
	return func(args map[string]any) bool {
		if statusUpdated, exists := args["StatusUpdatedEvent"]; exists {
			log.Printf("StubService methodCalled. %+v", args)
			return statusUpdated == status
		}
		return false
	}
}

func (s *StubService) RequestNew(ctx context.Context, itemName string, quantity int) error {
	args := map[string]interface{}{
		"itemName": itemName,
		"quantity": quantity,
	}
	s.MethodCalled = append(s.MethodCalled, args)
	return nil
}

func (s *StubService) EmitStatusUpdatedEvent(order *Order) {
	args := map[string]interface{}{
		"StatusUpdatedEvent": order.Status,
	}
	s.MethodCalled = append(s.MethodCalled, args)
}
