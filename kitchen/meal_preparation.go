package kitchen

import "time"

var preparationTimes = map[string]time.Duration{
	"hamburger":       2500,
	"cheeseburger":    4500,
	"double-cheese":   3750,
	"mc-chicken":      4200,
	"mr-chicken-wrap": 6000,
	"spicy-stripes":   4100,
	"hot-wings":       3200,
	"fries":           1500,
}

type MealPreparation interface {
	Prepare(item string, quantity int)
}

type MealPreparationService struct {
}

func (m *MealPreparationService) Prepare(item string, quantity int) {
	if duration, ok := preparationTimes[item]; ok {
		time.Sleep(duration * time.Duration(quantity))
		return
	}

	time.Sleep(time.Second)
}
