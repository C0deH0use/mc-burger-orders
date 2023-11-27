package stubs

import (
	"context"
	"encoding/json"
	m "mc-burger-orders/order"
)

type StubRepository struct {
	o              []*m.Order
	insertOrUpdate *m.Order
	fetchById      *m.Order
	nextNumber     int64
	err            error
	methodCalled   []map[string]interface{}
}

func GivenRepository() *StubRepository {
	return &StubRepository{}
}

func (s *StubRepository) ReturnFetchById(order *m.Order) {
	s.fetchById = order
}

func (s *StubRepository) ReturnOrders(orders ...*m.Order) {
	s.o = orders
}

func (s *StubRepository) ReturnWhenInsertOrUpdate(order *m.Order) {
	s.insertOrUpdate = order
}

func (s *StubRepository) ReturnNextNumber(nextNumber int64) {
	s.nextNumber = nextNumber
}

func (s *StubRepository) ReturnError(error error) {
	s.err = error
}

func (s *StubRepository) InsertOrUpdate(ctx context.Context, order m.Order) (*m.Order, error) {
	s.methodCalled = append(s.methodCalled, map[string]interface{}{"InsertOrUpdate": order})

	if s.insertOrUpdate != nil {
		return s.insertOrUpdate, nil
	}
	return &order, s.err
}

func (s *StubRepository) FetchById(ctx context.Context, id interface{}) (*m.Order, error) {
	s.methodCalled = append(s.methodCalled, map[string]interface{}{"FetchById": id})

	if s.fetchById != nil {
		return s.fetchById, nil
	}
	return s.o[0], s.err
}

func (s *StubRepository) FetchByOrderNumber(ctx context.Context, orderNumber int64) (*m.Order, error) {
	s.methodCalled = append(s.methodCalled, map[string]interface{}{"FetchByOrderNumber": orderNumber})
	return s.o[0], s.err
}

func (s *StubRepository) FetchMany(ctx context.Context) ([]*m.Order, error) {
	s.methodCalled = append(s.methodCalled, map[string]interface{}{"FetchMany": nil})
	return s.o, s.err
}

func (s *StubRepository) FetchByMissingItem(ctx context.Context, itemName string) ([]*m.Order, error) {
	s.methodCalled = append(s.methodCalled, map[string]interface{}{"FetchByMissingItem": itemName})
	return s.o, s.err
}

func (s *StubRepository) GetNext(ctx context.Context) (int64, error) {
	return s.nextNumber, s.err
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
