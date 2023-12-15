package management

import (
	"context"
	"github.com/segmentio/kafka-go"
	"mc-burger-orders/command"
	"mc-burger-orders/log"
	"mc-burger-orders/order"
)

type CheckMissingItemsOnOrdersCommand struct {
	queryService   OrderQueryService
	kitchenService order.KitchenRequestService
}

func (c *CheckMissingItemsOnOrdersCommand) Execute(ctx context.Context, message kafka.Message, result chan command.TypedResult) {
	log.Info.Printf("Checking for orders that where requested but items are still missing...")

	orders, err := c.queryService.FetchOrdersForPacking(ctx)
	if err != nil {
		result <- command.NewErrorResult("CheckMissingItemsOnOrdersCommand", err)
		return
	}

	for _, foundOrder := range orders {
		for _, missingItem := range foundOrder.GetMissingItems() {
			log.Info.Printf("Order %d, is missing %v in quantity - %d", foundOrder.OrderNumber, missingItem.Name, missingItem.Quantity)

			err = c.kitchenService.RequestNew(ctx, missingItem.Name, missingItem.Quantity)
			if err != nil {
				result <- command.NewErrorResult("CheckMissingItemsOnOrdersCommand", err)
			}
		}
	}

	result <- command.NewSuccessfulResult("CheckMissingItemsOnOrdersCommand")
}
