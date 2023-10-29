package model

import "mc-burger-orders/item"

type NewOrder struct {
	CustomerId int16       `json:"customerId" binding:"required"`
	Items      []item.Item `json:"items" binding:"required,gt=0"`
}
