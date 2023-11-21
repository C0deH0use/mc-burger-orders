package order

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/segmentio/kafka-go"
	"mc-burger-orders/log"
	"mc-burger-orders/order/dto"
	om "mc-burger-orders/order/model"
	"mc-burger-orders/order/service"
	"mc-burger-orders/stack"
)

type PackItemCommand struct {
	Repository     om.PackingOrderItemsRepository
	KitchenService service.KitchenRequestService
	StatusEmitter  StatusEmitter
	Stack          *stack.Stack
	Message        kafka.Message
}

func (p *PackItemCommand) Execute(ctx context.Context) (bool, error) {
	stackMessage := make([]dto.StackItemAddedMessage, 0)
	err := json.Unmarshal(p.Message.Value, &stackMessage)
	if err != nil {
		log.Error.Println("could not Unmarshal event message to StackUpdatedMessage", err)
		return false, err
	}

	if len(stackMessage) == 0 {
		err := fmt.Errorf("event message is nil or empty")
		return false, err
	}

	for _, itemUpdate := range stackMessage {
		log.Info.Printf("New item(s) %v added to stack", itemUpdate.ItemName)

		orders, err := p.Repository.FetchByMissingItem(ctx, itemUpdate.ItemName)
		if err != nil {
			log.Error.Println("could not find orders with any items still not packed", err)
			return false, err
		}

		for _, order := range orders {
			orderQuantity, err := getMissingItemsOfOrder(order, itemUpdate.ItemName)
			if err != nil {
				log.Error.Println(err.Error())
				continue
			}

			current := p.Stack.GetCurrent(itemUpdate.ItemName)
			if current < orderQuantity {
				quantityToRequest := orderQuantity - current
				err = p.KitchenService.RequestForOrder(ctx, itemUpdate.ItemName, quantityToRequest, order.OrderNumber)

				if err != nil {
					log.Error.Println(err.Error())
					return false, err
				}

				orderQuantity = current
			}

			err = p.Stack.Take(itemUpdate.ItemName, orderQuantity)
			if err != nil {
				log.Error.Printf("could not take item `%v` in quantity `%d` from Stack. Reason: %v", itemUpdate.ItemName, itemUpdate.Quantity, err)
				continue
			}

			statusUpdated := order.PackItem(itemUpdate.ItemName, orderQuantity)
			_, err = p.Repository.InsertOrUpdate(ctx, *order)
			if err != nil {
				log.Error.Printf("failed to update order `%d`, reason: %v", order.OrderNumber, err)
				return false, err
			}
			if statusUpdated {
				go p.StatusEmitter.EmitStatusUpdatedEvent(order)
			}
		}
	}

	return true, nil
}

func getMissingItemsOfOrder(o *om.Order, itemName string) (int, error) {
	var quantity = -1
	for _, item := range o.Items {
		if item.Name == itemName {
			quantity = item.Quantity
		}
	}
	if quantity == -1 {
		err := fmt.Errorf("could not find item `%v` on the order list", itemName)
		return -quantity, err
	}

	for _, item := range o.PackedItems {
		if item.Name == itemName {
			quantity -= item.Quantity
		}
	}

	return quantity, nil
}
