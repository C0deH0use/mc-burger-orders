package order

import (
	"context"
	"encoding/json"
	m "mc-burger-orders/order/model"
)

type StubRepository struct {
	o            *m.Order
	nextNumber   int64
	err          error
	methodCalled []map[string]interface{}
}

func (s *StubRepository) InsertOrUpdate(ctx context.Context, order m.Order) (*m.Order, error) {
	s.methodCalled = append(s.methodCalled, map[string]interface{}{"InsertOrUpdate": order})
	return s.o, s.err
}
func (s *StubRepository) FetchById(ctx context.Context, id interface{}) (*m.Order, error) {
	s.methodCalled = append(s.methodCalled, map[string]interface{}{"FetchById": id})
	return s.o, nil
}
func (s *StubRepository) FetchByOrderNumber(ctx context.Context, orderNumber int64) (*m.Order, error) {
	s.methodCalled = append(s.methodCalled, map[string]interface{}{"FetchByOrderNumber": orderNumber})
	return s.o, nil
}
func (s *StubRepository) FetchMany(ctx context.Context) ([]m.Order, error) {
	s.methodCalled = append(s.methodCalled, map[string]interface{}{"FetchMany": nil})
	return []m.Order{*s.o}, nil
}

func (e *StubRepository) GetNext(ctx context.Context) (int64, error) {
	return e.nextNumber, nil
}

func (s *StubRepository) GetUpsertArgs() []m.Order {
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

func (s *StubRepository) CalledCnt() int {
	return len(s.methodCalled)
}

func (s *StubRepository) HaveBeenCalledWith(itemName string, quantity int, orderNumber int64) bool {
	r := false

	for _, args := range s.methodCalled {
		argName := args["itemName"]
		argQuantity := args["quantity"]
		argNumber := args["orderNumber"]
		if argName == itemName && argQuantity == quantity && argNumber == orderNumber {
			r = true
		}
	}

	return r
}
