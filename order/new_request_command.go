package order

import (
	"context"
	"fmt"
	item2 "mc-burger-orders/kitchen/item"
	"mc-burger-orders/log"
	om "mc-burger-orders/order/model"
	"mc-burger-orders/order/service"
	"mc-burger-orders/stack"
)

type NewRequestCommand struct {
	Repository     om.OrderRepository
	Stack          *stack.Stack
	KitchenService service.KitchenRequestService
	StatusEmitter  StatusEmitter
	OrderNumber    int64
	NewOrder       om.NewOrder
}

func (c *NewRequestCommand) Execute(ctx context.Context) (bool, error) {
	orderRecord := om.CreateNewOrder(c.OrderNumber, c.NewOrder)

	log.Info.Printf("New Order with number %v created %+v\n", c.OrderNumber, c.NewOrder)
	statusUpdated := false
	for _, item := range c.NewOrder.Items {
		isReady, err := item2.IsItemReady(item.Name)
		if err != nil {
			return false, err
		}

		if isReady {
			log.Info.Printf("Item %v is of type automatically ready. No need to check stack if one in available. Packing automatically.", item.Name)
			if sUpdated := orderRecord.PackItem(item.Name, item.Quantity); sUpdated {
				statusUpdated = sUpdated
			}
		} else {
			sUpdated, err := c.handlePreparationItems(ctx, item, &orderRecord)
			if err != nil {
				return false, err
			}
			if sUpdated {
				statusUpdated = sUpdated
			}
		}
	}

	result, err := c.Repository.InsertOrUpdate(ctx, orderRecord)
	if err != nil {
		return false, err
	}
	if result == nil {
		return false, fmt.Errorf("failed to store Order in DB, despite MongoDB Driver returning success")
	}
	if statusUpdated {
		go c.StatusEmitter.EmitStatusUpdatedEvent(result)
	}
	return true, nil
}

func (c *NewRequestCommand) handlePreparationItems(ctx context.Context, item item2.Item, orderRecord *om.Order) (statusUpdated bool, err error) {
	log.Info.Println("Item", item, "needs to be prepared first. Checking stack if one in available.")
	amountInStock := c.Stack.GetCurrent(item.Name)
	if amountInStock == 0 {
		log.Info.Printf("Sending Request to kitchen for %d new %v", item.Quantity, item.Name)
		err = c.KitchenService.RequestForOrder(ctx, item.Name, item.Quantity, orderRecord.OrderNumber)
		if err != nil {
			return statusUpdated, err
		}
	} else {
		var itemTaken int

		if amountInStock > item.Quantity {
			log.Info.Println("Stack has required quantity of item", item.Name)
			itemTaken = item.Quantity
		} else {
			itemTaken = item.Quantity - amountInStock
			remaining := item.Quantity - itemTaken

			log.Info.Printf("Sending Request to kitchen for %d new %v", remaining, item.Name)
			err = c.KitchenService.RequestForOrder(ctx, item.Name, remaining, orderRecord.OrderNumber)
			if err != nil {
				return statusUpdated, err
			}
		}
		err = c.Stack.Take(item.Name, itemTaken)

		if err != nil {
			err = fmt.Errorf("error when collecting '%d' item(s) '%s' from stack", item.Quantity, item.Name)
			return statusUpdated, err
		}

		log.Info.Printf("Packing %d of %v into order %d", itemTaken, item.Name, orderRecord.OrderNumber)
		statusUpdated = orderRecord.PackItem(item.Name, itemTaken)
	}
	return statusUpdated, err
}
