package kitchen

import (
	"mc-burger-orders/log"
)

func (h *Handler) CreateNewItem(request ItemRequest) (bool, error) {
	log.Info.Printf("Cook | Starting to prepare new item -> %v in amount: `%d`", request.ItemName, request.Quantity)
	//

	log.Info.Printf("Cook | item(s) %v in prepared", 1, request.ItemName)

	//
	log.Info.Printf("Cook | item(s) %v added to stack", 1, request.ItemName)
	return false, nil
}
