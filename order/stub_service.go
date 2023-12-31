package order

import (
	"context"
	"log"
	"mc-burger-orders/testing/stubs"
	"sync"
)

type StubService struct {
	wg *sync.WaitGroup
	stubs.DefaultStubService
}

func NewStubService() *StubService {
	return &StubService{wg: nil, DefaultStubService: stubs.DefaultStubService{MethodCalled: make([]map[string]any, 0)}}
}

func (s *StubService) WithWaitGroup(group *sync.WaitGroup) {
	s.wg = group
}

func RequestMatchingFnc(itemName string, quantity int) func(args map[string]any) bool {
	return func(args map[string]any) bool {
		argName := args["itemName"]
		argQuantity := args["quantity"]
		b := argName == itemName && argQuantity == quantity
		log.Printf("StubService methodCalled RequestNew() => %+v", args)
		return b
	}
}

func StatusUpdateMatchingFnc(status OrderStatus) func(args map[string]any) bool {
	return func(args map[string]any) bool {
		if updatedStatus, exists := args["StatusUpdatedEvent"]; exists {
			log.Printf("StubService methodCalled StatusUpdatedEvent() => %+v", updatedStatus)
			return updatedStatus == status
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

	if s.wg != nil {
		s.wg.Done()
	}
	return nil
}

func (s *StubService) EmitStatusUpdatedEvent(o Order) {
	orderCopy := &o
	args := map[string]interface{}{
		"StatusUpdatedEvent": orderCopy.Status,
	}
	s.MethodCalled = append(s.MethodCalled, args)
	if s.wg != nil {
		s.wg.Done()
	}
	log.Printf("StubService.EmitStatusUpdatedEvent() => %+v", args)
}

func (s *StubService) EmitUpdatedEvent(o Order) {
	orderCopy := &o
	args := map[string]interface{}{
		"EmitUpdatedEvent": orderCopy,
	}
	s.MethodCalled = append(s.MethodCalled, args)
	if s.wg != nil {
		s.wg.Done()
	}
}

func (s *StubService) GetStatusUpdatedEventArgs() []OrderStatus {
	r := make([]OrderStatus, 0)

	for _, o := range s.MethodCalled {
		if value, eventExist := o["StatusUpdatedEvent"]; eventExist {
			r = append(r, value.(OrderStatus))
		}
	}
	return r
}

func (s *StubService) GetEmitUpdatedEventArgs() []*Order {
	r := make([]*Order, 0)

	for _, o := range s.MethodCalled {
		if value, eventExist := o["EmitUpdatedEvent"]; eventExist {
			order := value.(*Order)
			r = append(r, order)
		}
	}
	return r
}
