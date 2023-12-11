package order

import (
	"context"
	"github.com/segmentio/kafka-go"
	"mc-burger-orders/command"
)

type OrderCollectedCommand struct {
	Repository  OrderRepository
	OrderNumber int64
}

func (o *OrderCollectedCommand) Execute(ctx context.Context, message kafka.Message, commandResults chan command.TypedResult) {
	//TODO implement me
	panic("implement me")
}
