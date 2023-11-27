package model

import (
	"mc-burger-orders/kitchen/item"
)

type NewOrder struct {
	CustomerId int         `json:"customerId" binding:"required"`
	Items      []item.Item `json:"items" binding:"required,gt=0"`
}
