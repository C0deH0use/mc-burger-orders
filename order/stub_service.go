package order

import (
	"context"
	"log"
	"mc-burger-orders/testing/stubs"
)

type StubService struct {
	stubs.DefaultStubService
}

func NewOrderService() *StubService {
	return &StubService{stubs.DefaultStubService{MethodCalled: make([]map[string]any, 0)}}
}

func RequestMatchingFnc(itemName string, quantity int, orderNumber int64) func(args map[string]any) bool {
	return func(args map[string]any) bool {
		argName := args["itemName"]
		argQuantity := args["quantity"]
		argNumber := args["orderNumber"]
		log.Printf("StubService methodCalled. %+v", args)
		b := argName == itemName && argQuantity == quantity && argNumber == orderNumber
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

func (s *StubService) RequestForOrder(ctx context.Context, itemName string, quantity int, orderNumber int64) error {
	args := map[string]interface{}{
		"itemName":    itemName,
		"quantity":    quantity,
		"orderNumber": orderNumber,
	}
	s.MethodCalled = append(s.MethodCalled, args)
	return nil
}

//func (s *StubService) SendMessage(ctx context.Context, messages ...kafka.Message) error {
//	args := map[string]interface{}{
//		"SendMessage": messages,
//	}
//	s.MethodCalled = append(s.MethodCalled, args)
//	return nil
//}

func (s *StubService) EmitStatusUpdatedEvent(order *Order) {
	args := map[string]interface{}{
		"StatusUpdatedEvent": order.Status,
	}
	s.MethodCalled = append(s.MethodCalled, args)
}
