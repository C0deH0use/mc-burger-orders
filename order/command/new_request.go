package command

import (
	"context"
	"fmt"
	i "mc-burger-orders/item"
	"mc-burger-orders/log"
	om "mc-burger-orders/order/model"
	"mc-burger-orders/order/service"
	"mc-burger-orders/stack"
)

type NewRequestCommand struct {
	Repository     om.OrderRepository
	Stack          *stack.Stack
	KitchenService service.KitchenRequestService
	OrderNumber    int64
	NewOrder       om.NewOrder
}

func (c *NewRequestCommand) Execute(ctx context.Context) (bool, error) {
	order := om.CreateNewOrder(c.OrderNumber, c.NewOrder)

	log.Info.Printf("New Order with number %v created %+v\n", c.OrderNumber, c.NewOrder)

	for _, item := range c.NewOrder.Items {
		isReady, err := i.IsItemReady(item.Name)
		if err != nil {
			return false, err
		}

		if isReady {
			log.Info.Printf("Item %v is of type automatically ready. No need to check stack if one in available. Packing automatically.", item.Name)
			order.PackItem(item.Name, item.Quantity)
		} else {
			err = c.handlePreparationItems(ctx, item, &order)
			if err != nil {
				return false, err
			}
		}
	}

	result, err := c.Repository.InsertOrUpdate(ctx, order)
	if err != nil {
		return false, err
	}
	if result == nil {
		return false, fmt.Errorf("failed to store Order in DB, despite MongoDB Driver returning success")
	}

	return true, nil
}

func (c *NewRequestCommand) handlePreparationItems(ctx context.Context, item i.Item, order *om.Order) (err error) {
	log.Info.Println("Item", item, "needs to be prepared first. Checking stack if one in available.")
	amountInStock := c.Stack.GetCurrent(item.Name)
	if amountInStock == 0 {
		log.Info.Printf("Sending Request to kitchen for %d new %v", item.Quantity, item.Name)
		err := c.KitchenService.RequestForOrder(ctx, item.Name, item.Quantity, order.OrderNumber)
		if err != nil {
			return err
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
			err := c.KitchenService.RequestForOrder(ctx, item.Name, remaining, order.OrderNumber)
			if err != nil {
				return err
			}
		}
		err = c.Stack.Take(item.Name, itemTaken)

		if err != nil {
			err = fmt.Errorf("error when collecting '%d' item(s) '%s' from stack", item.Quantity, item.Name)
			return err
		}

		log.Info.Printf("Packing %d of %v into order %d", itemTaken, item.Name, order.OrderNumber)
		order.PackItem(item.Name, itemTaken)
	}
	return nil
}
