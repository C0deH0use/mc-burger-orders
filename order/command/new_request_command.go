package command

import (
	"context"
	"fmt"
	"log"
	i "mc-burger-orders/item"
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

	for _, item := range c.NewOrder.Items {
		isReady, err := i.IsItemReady(item.Name)
		if err != nil {
			return false, err
		}

		if isReady {
			order.PackItem(item.Name, item.Quantity)
		} else {
			err = c.handlePreparationItems(item, &order)
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

func (c *NewRequestCommand) handlePreparationItems(item i.Item, order *om.Order) (err error) {
	amountInStock := c.Stack.GetCurrent(item.Name)
	if amountInStock == 0 {
		c.KitchenService.Request(item.Name, item.Quantity)
		log.Println("Sending Request to kitchen for", item.Quantity, "new", item.Name)
	} else {
		var itemTaken int
		if amountInStock > item.Quantity {
			itemTaken = item.Quantity
		} else {
			itemTaken = item.Quantity - amountInStock
			remaining := item.Quantity - itemTaken
			log.Println("Sending Request to kitchen for", remaining, "new", item.Name)
			c.KitchenService.Request(item.Name, item.Quantity)
		}
		err = c.Stack.Take(item.Name, itemTaken)

		if err != nil {
			err = fmt.Errorf("error when collecting '%d' item(s) '%s' from stack", item.Quantity, item.Name)
			return err
		}

		order.PackItem(item.Name, itemTaken)
	}
	return nil
}
