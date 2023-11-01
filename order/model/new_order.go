package model

import "mc-burger-orders/item"

type NewOrder struct {
	CustomerId int         `json:"customerId" binding:"required"`
	Items      []item.Item `json:"items" binding:"required,gt=0"`
}
