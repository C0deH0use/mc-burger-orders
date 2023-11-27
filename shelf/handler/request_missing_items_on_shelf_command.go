package handler

import (
	"context"
	"github.com/segmentio/kafka-go"
	"mc-burger-orders/shelf"
)

type RequestMissingItemsOnShelfCommand struct {
	KitchenService KitchenService
	Shelf          *shelf.Shelf
}

func (r *RequestMissingItemsOnShelfCommand) Execute(ctx context.Context, message kafka.Message) (bool, error) {
	//TODO implement me
	return false, nil
}
