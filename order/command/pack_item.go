package command

import (
	"context"
	om "mc-burger-orders/order/model"
	"mc-burger-orders/stack"
)

type PackItemCommand struct {
	Repository  om.OrderRepository
	Stack       *stack.Stack
	OrderNumber int64
}

func (p *PackItemCommand) Execute(ctx context.Context) (bool, error) {
	//TODO implement me
	panic("implement me")
}
