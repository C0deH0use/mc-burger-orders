package order

import (
	"context"
	om "mc-burger-orders/order/model"
)

type OrderCollectedCommand struct {
	Repository  om.OrderRepository
	OrderNumber int64
}

func (o *OrderCollectedCommand) Execute(ctx context.Context) (bool, error) {
	//TODO implement me
	panic("implement me")
}
