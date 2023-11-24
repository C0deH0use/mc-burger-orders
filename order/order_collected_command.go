package order

import (
	"context"
	"github.com/segmentio/kafka-go"
	om "mc-burger-orders/order/model"
)

type OrderCollectedCommand struct {
	Repository  om.OrderRepository
	OrderNumber int64
}

func (o *OrderCollectedCommand) Execute(ctx context.Context, message kafka.Message) (bool, error) {
	//TODO implement me
	panic("implement me")
}
