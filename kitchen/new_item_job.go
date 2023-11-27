package kitchen

import (
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"mc-burger-orders/log"
	"mc-burger-orders/order/utils"
)

func (h *Handler) CreateNewItem(message kafka.Message) (bool, error) {
	requests := &[]ItemRequest{}
	err := json.Unmarshal(message.Value, requests)
	if err != nil {
		log.Error.Println(err.Error())
		return false, err
	}
	orderNumber, err := utils.GetOrderNumber(message)
	if err != nil {
		orderNumber = -1
	}

	messageKey := string(message.Key)
	log.Info.Printf("CookRequest: %v | Order: %d", messageKey, orderNumber)

	for _, request := range *requests {
		log.Info.Printf("CookRequest: %v | Order: %d | Starting to prepare new item -> %v in amount: `%d`", messageKey, orderNumber, request.ItemName, request.Quantity)

		h.mealPreparation.Prepare(request.ItemName, request.Quantity)

		log.Info.Printf("CookRequest: %v | Order: %d | item(s) %v in prepared", messageKey, orderNumber, request.ItemName)

		h.stack.AddMany(request.ItemName, request.Quantity)
	}

	return true, nil
}
