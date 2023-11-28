package item

import "time"

type Item struct {
	Name     string `json:"name" binding:"required"`
	Quantity int    `json:"quantity" binding:"gt=0"`
}

type MenuItemConfigs struct {
	InstantReady    bool
	Favorite        bool
	PreparationTime time.Duration
}

var MenuItems = map[string]MenuItemConfigs{
	"hamburger":       {false, true, 2500},
	"cheeseburger":    {false, true, 4500},
	"double-cheese":   {false, false, 3750},
	"mc-spicy":        {false, false, 3200},
	"mc-chicken":      {false, false, 4200},
	"mr-chicken-wrap": {false, false, 6000},
	"spicy-stripes":   {false, true, 4100},
	"hot-wings":       {false, true, 3200},
	"fries":           {false, true, 1500},
	"coke":            {true, false, 0},
	"ice-cream":       {true, false, 0},
	"fanta":           {true, false, 0},
}
