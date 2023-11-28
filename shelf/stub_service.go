package shelf

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

func NewShelfStubService() *StubService {
	return &StubService{nil, stubs.DefaultStubService{MethodCalled: make([]map[string]any, 0)}}
}

func NewShelfStubServiceWithWG(wg *sync.WaitGroup) *StubService {
	return &StubService{wg, stubs.DefaultStubService{MethodCalled: make([]map[string]any, 0)}}
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
