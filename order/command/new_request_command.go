package command

import (
	"fmt"
	i "mc-burger-orders/item"
	om "mc-burger-orders/order/model"
	"mc-burger-orders/order/service"
	"mc-burger-orders/stack"
)

type NewRequestCommand struct {
	Repository     om.OrderRepository
	Stack          *stack.Stack
	KitchenService service.KitchenRequestService
	NewOrder       om.NewOrder
}

func (c *NewRequestCommand) Execute() (*om.Order, error) {
	order := c.Repository.CreateOrder(c.NewOrder)
	for _, item := range c.NewOrder.Items {
		isReady, err := i.IsItemReady(item.Name)
		if err != nil {
			return nil, err
		}

		if isReady {
			order.PackItem(item, item.Quantity)
		} else {
			err = c.handlePreparationItems(item, order)
			if err != nil {
				return nil, err
			}
		}
	}

	return order, nil
}

func (c *NewRequestCommand) handlePreparationItems(item i.Item, order *om.Order) (err error) {
	amountInStock := c.Stack.GetCurrent(item.Name)
	if amountInStock == 0 {
		c.KitchenService.Request(item.Name, item.Quantity)
		fmt.Println("Sending Request to kitchen for", item.Quantity, "new", item.Name)
	} else {
		var itemTaken int
		if amountInStock > item.Quantity {
			itemTaken = item.Quantity
		} else {
			itemTaken = item.Quantity - amountInStock
			remaining := item.Quantity - itemTaken
			fmt.Println("Sending Request to kitchen for", remaining, "new", item.Name)
			c.KitchenService.Request(item.Name, item.Quantity)
		}
		err = c.Stack.Take(item.Name, itemTaken)

		if err != nil {
			err = fmt.Errorf("error when collecting '%d' item(s) '%s' from stack", item.Quantity, item.Name)
			return err
		}

		order.PackItem(item, itemTaken)
	}
	return nil
}
