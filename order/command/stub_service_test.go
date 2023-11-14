package command

import (
	"context"
	"log"
	m "mc-burger-orders/order/model"
)

type StubService struct {
	o            *m.Order
	err          error
	methodCalled []map[string]interface{}
}

func (s *StubService) CalledCnt() int {
	return len(s.methodCalled)
}

func (s *StubService) HaveBeenCalledWith(itemName string, quantity int, orderNumber int64) bool {
	r := false

	for _, args := range s.methodCalled {
		argName := args["itemName"]
		argQuantity := args["quantity"]
		argNumber := args["orderNumber"]
		log.Printf("StubService methodCalled. %+v", args)
		if argName == itemName && argQuantity == quantity && argNumber == orderNumber {
			r = true
		}
	}

	return r
}

func (s *StubService) RequestForOrder(ctx context.Context, itemName string, quantity int, orderNumber int64) error {
	args := map[string]interface{}{
		"itemName":    itemName,
		"quantity":    quantity,
		"orderNumber": orderNumber,
	}
	s.methodCalled = append(s.methodCalled, args)
	return nil
}
