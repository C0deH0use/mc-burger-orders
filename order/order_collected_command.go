package order

import (
	"context"
	"github.com/segmentio/kafka-go"
)

type OrderCollectedCommand struct {
	Repository  OrderRepository
	OrderNumber int64
}

func (o *OrderCollectedCommand) Execute(ctx context.Context, message kafka.Message) (bool, error) {
	//TODO implement me
	panic("implement me")
}
