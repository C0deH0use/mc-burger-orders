package kitchen

import (
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"mc-burger-orders/log"
)

func (h *Handler) CreateNewItem(message kafka.Message) (bool, error) {
	requests := &[]ItemRequest{}
	err := json.Unmarshal(message.Value, requests)
	if err != nil {
		log.Error.Println(err.Error())
		return false, err
	}

	messageKey := string(message.Key)
	log.Info.Printf("CookRequest: %v | Order: %d", messageKey)

	for _, request := range *requests {
		log.Info.Printf("CookRequest: %v | Starting to prepare new item -> %v in amount: `%d`", messageKey, request.ItemName, request.Quantity)

		h.mealPreparation.Prepare(request.ItemName, request.Quantity)

		log.Info.Printf("CookRequest: %v | item(s) %v in prepared", messageKey, request.ItemName)

		h.shelf.AddMany(request.ItemName, request.Quantity)
	}

	return true, nil
}
