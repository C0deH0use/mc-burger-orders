package kitchen

import (
	"log"
	"mc-burger-orders/testing/stubs"
)

type StubService struct {
	stubs.DefaultStubService
}

func NewMealPrepService() *StubService {
	return &StubService{stubs.DefaultStubService{MethodCalled: make([]map[string]any, 0)}}
}

func (s *StubService) Prepare(item string, quantity int) {
	args := map[string]interface{}{
		"itemName": item,
		"quantity": quantity,
	}
	s.MethodCalled = append(s.MethodCalled, args)
}

func MealPrepMatchingFnc(itemName string, quantity int) func(args map[string]any) bool {
	return func(args map[string]any) bool {
		argName := args["itemName"]
		argQuantity := args["quantity"]
		log.Printf("StubService methodCalled. %+v", args)
		return argName == itemName && argQuantity == quantity
	}
}
