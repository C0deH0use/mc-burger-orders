package kitchen

import (
	item2 "mc-burger-orders/kitchen/item"
	"time"
)

type MealPreparation interface {
	Prepare(item string, quantity int)
}

type MealPreparationService struct {
}

func (m *MealPreparationService) Prepare(item string, quantity int) {
	if itemConfig, ok := item2.MenuItems[item]; ok {
		time.Sleep(itemConfig.PreparationTime * time.Duration(quantity))
		return
	}

	time.Sleep(time.Second)
}
