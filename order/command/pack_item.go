package command

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/segmentio/kafka-go"
	"mc-burger-orders/log"
	"mc-burger-orders/order/dto"
	om "mc-burger-orders/order/model"
	"mc-burger-orders/order/service"
	"mc-burger-orders/order/utils"
	"mc-burger-orders/stack"
)

type PackItemCommand struct {
	Repository     om.OrderRepository
	KitchenService service.KitchenRequestService
	Stack          *stack.Stack
	Message        kafka.Message
}

func (p *PackItemCommand) Execute(ctx context.Context) (bool, error) {
	orderNumber, err := utils.GetOrderNumber(p.Message)
	if err != nil {
		log.Error.Println("could not find order number from message", err)
		return false, err
	}

	log.Info.Printf("New item added to stack. Order %d can proceed packing", orderNumber)

	order, err := p.Repository.FetchByOrderNumber(ctx, orderNumber)
	if err != nil {
		log.Error.Println("could not find order by number", err)
		return false, err
	}

	stackMessage := dto.StackUpdatedMessage{}
	err = json.Unmarshal(p.Message.Value, &stackMessage)
	if err != nil {
		log.Error.Println("could not Unmarshal event message to StackUpdatedMessage", err)
		return false, err
	}

	if len(stackMessage) == 0 {
		err := fmt.Errorf("event message is nil or empty")
		return false, err
	}

	for _, itemUpdate := range stackMessage {
		orderQuantity, err := getOrderQuantityForItem(order, itemUpdate.ItemName)
		if err != nil {
			log.Error.Println(err.Error())
			return false, err
		}

		current := p.Stack.GetCurrent(itemUpdate.ItemName)
		if current < orderQuantity {
			quantityToRequest := orderQuantity - current
			err = p.KitchenService.RequestForOrder(ctx, itemUpdate.ItemName, quantityToRequest, orderNumber)

			if err != nil {
				log.Error.Println(err.Error())
				return false, err
			}

			orderQuantity = current
		}

		err = p.Stack.Take(itemUpdate.ItemName, orderQuantity)
		if err != nil {
			log.Error.Printf("could not take item `%v` in quantity `%d` from Stack. Reason: %v", itemUpdate.ItemName, itemUpdate.Quantity, err)
			return false, err
		}
		order.PackItem(itemUpdate.ItemName, orderQuantity)
	}

	_, err = p.Repository.InsertOrUpdate(ctx, *order)
	if err != nil {
		log.Error.Printf("failed to update order `%d`, reason: %v", orderNumber, err)
		return false, err
	}

	return true, nil
}

func getOrderQuantityForItem(o *om.Order, itemName string) (int, error) {
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
