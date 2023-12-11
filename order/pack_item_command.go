package order

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/segmentio/kafka-go"
	"mc-burger-orders/command"
	"mc-burger-orders/log"
	"mc-burger-orders/order/dto"
	"mc-burger-orders/shelf"
)

type PackItemCommand struct {
	Repository     PackingOrderItemsRepository
	KitchenService KitchenRequestService
	StatusEmitter  StatusEmitter
	Shelf          *shelf.Shelf
}

func (p *PackItemCommand) Execute(ctx context.Context, message kafka.Message, commandResults chan command.TypedResult) {
	stackMessage := make([]dto.StackItemAddedMessage, 0)
	err := json.Unmarshal(message.Value, &stackMessage)
	if err != nil {
		log.Error.Println("could not Unmarshal event message to StackUpdatedMessage", err)
		commandResults <- command.NewErrorResult("PackItemCommand", err)
		return
	}

	if len(stackMessage) == 0 {
		err := fmt.Errorf("event message is nil or empty")
		commandResults <- command.NewErrorResult("PackItemCommand", err)
		return
	}

	for _, itemUpdate := range stackMessage {
		log.Info.Printf("New item(s) %v added to shelf", itemUpdate.ItemName)

		orders, err := p.Repository.FetchByMissingItem(ctx, itemUpdate.ItemName)
		if err != nil {
			log.Error.Println("could not find orders with any items still not packed", err)
			commandResults <- command.NewErrorResult("PackItemCommand", err)
			return
		}

		for _, order := range orders {
			orderQuantity, err := order.GetMissingItemsCount(itemUpdate.ItemName)
			if err != nil {
				log.Error.Println(err.Error())
				continue
			}

			current := p.Shelf.GetCurrent(itemUpdate.ItemName)
			if current < orderQuantity {
				quantityToRequest := orderQuantity - current
				err = p.KitchenService.RequestNew(ctx, itemUpdate.ItemName, quantityToRequest)

				if err != nil {
					commandResults <- command.NewErrorResult("PackItemCommand", err)
					return
				}

				orderQuantity = current
			}

			err = p.Shelf.Take(itemUpdate.ItemName, orderQuantity)
			if err != nil {
				log.Error.Printf("could not take item `%v` in quantity `%d` from Shelf. Reason: %v", itemUpdate.ItemName, itemUpdate.Quantity, err)
				continue
			}

			statusUpdated := order.PackItem(itemUpdate.ItemName, orderQuantity)
			_, err = p.Repository.InsertOrUpdate(ctx, *order)
			if err != nil {
				log.Error.Printf("failed to update order `%d`, reason: %v", order.OrderNumber, err)
				commandResults <- command.NewErrorResult("PackItemCommand", err)
				return
			}
			if statusUpdated {
				p.StatusEmitter.EmitStatusUpdatedEvent(order)
			}
		}
	}

	commandResults <- command.NewSuccessfulResult("PackItemCommand")
}
