package order

import (
	"context"
	"fmt"
	"github.com/segmentio/kafka-go"
	"mc-burger-orders/command"
	item2 "mc-burger-orders/kitchen/item"
	"mc-burger-orders/log"
	"mc-burger-orders/shelf"
)

type NewRequestCommand struct {
	Repository     OrderRepository
	Stack          *shelf.Shelf
	KitchenService KitchenRequestService
	StatusEmitter  StatusEmitter
	OrderNumber    int64
	NewOrder       NewOrder
}

func (c *NewRequestCommand) Execute(ctx context.Context, message kafka.Message, commandResults chan command.TypedResult) {
	orderRecord := CreateNewOrder(c.OrderNumber, c.NewOrder)

	log.Info.Printf("New Order with number %v created %+v\n", c.OrderNumber, c.NewOrder)
	statusUpdated := false
	for _, item := range c.NewOrder.Items {
		isReady, err := item2.IsItemReady(item.Name)
		if err != nil {
			commandResults <- command.NewErrorResult("NewRequestCommand", err)
			return
		}

		if isReady {
			log.Info.Printf("Item %v is of type automatically ready. No need to check shelf if one in available. Packing automatically.", item.Name)
			if sUpdated := orderRecord.PackItem(item.Name, item.Quantity); sUpdated {
				statusUpdated = sUpdated
			}
		} else {
			sUpdated, err := c.handlePreparationItems(ctx, item, &orderRecord)
			if err != nil {
				commandResults <- command.NewErrorResult("NewRequestCommand", err)
				return
			}
			if sUpdated {
				statusUpdated = sUpdated
			}
		}
	}

	result, err := c.Repository.InsertOrUpdate(ctx, orderRecord)
	if err != nil {
		commandResults <- command.NewErrorResult("NewRequestCommand", err)
		return
	}
	if result == nil {
		commandResults <- command.NewErrorResult("NewRequestCommand", fmt.Errorf("failed to store Order in DB, despite MongoDB Driver returning success"))
		return
	}
	if statusUpdated {
		c.StatusEmitter.EmitStatusUpdatedEvent(result)
	}
	commandResults <- command.NewSuccessfulResult("NewRequestCommand")
}

func (c *NewRequestCommand) handlePreparationItems(ctx context.Context, item item2.Item, orderRecord *Order) (statusUpdated bool, err error) {
	log.Info.Println("Item", item, "needs to be prepared first. Checking shelf if one in available.")
	amountInStock := c.Stack.GetCurrent(item.Name)
	if amountInStock == 0 {
		log.Info.Printf("Sending Request to kitchen for %d new %v", item.Quantity, item.Name)
		err = c.KitchenService.RequestNew(ctx, item.Name, item.Quantity)
		if err != nil {
			return statusUpdated, err
		}
	} else {
		var itemTaken int

		if amountInStock > item.Quantity {
			log.Info.Println("Shelf has required quantity of item", item.Name)
			itemTaken = item.Quantity
		} else {
			itemTaken = item.Quantity - amountInStock
			remaining := item.Quantity - itemTaken

			log.Info.Printf("Sending Request to kitchen for %d new %v", remaining, item.Name)
			err = c.KitchenService.RequestNew(ctx, item.Name, remaining)
			if err != nil {
				return statusUpdated, err
			}
		}
		err = c.Stack.Take(item.Name, itemTaken)

		if err != nil {
			err = fmt.Errorf("error when collecting '%d' item(s) '%s' from shelf", item.Quantity, item.Name)
			return statusUpdated, err
		}

		log.Info.Printf("Packing %d of %v into order %d", itemTaken, item.Name, orderRecord.OrderNumber)
		statusUpdated = orderRecord.PackItem(item.Name, itemTaken)
	}
	return statusUpdated, err
}
