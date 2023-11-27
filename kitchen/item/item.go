package item

import "time"

type Item struct {
	Name     string `json:"name" binding:"required"`
	Quantity int    `json:"quantity" binding:"gt=0"`
}

type MenuItemConfigs struct {
	InstantReady    bool
	PreparationTime time.Duration
}

var MenuItems = map[string]MenuItemConfigs{
	"hamburger":       {false, 2500},
	"cheeseburger":    {false, 4500},
	"double-cheese":   {false, 3750},
	"mc-spicy":        {false, 3200},
	"mc-chicken":      {false, 4200},
	"mr-chicken-wrap": {false, 6000},
	"spicy-stripes":   {false, 4100},
	"hot-wings":       {false, 3200},
	"fries":           {false, 1500},
	"coke":            {true, 0},
	"ice-cream":       {true, 0},
	"fanta":           {true, 0},
}
