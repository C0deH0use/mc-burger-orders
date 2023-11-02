package model

import (
	"mc-burger-orders/item"
	"time"
)

type Order struct {
	OrderNumber int         `json:"orderNumber"`
	CustomerId  int         `json:"customerId"`
	Items       []item.Item `json:"items"`
	PackedItems []item.Item `json:"packedItems"`
	Status      OrderStatus `json:"status" bson:"status"`
	CreatedAt   time.Time   `json:"createdAt"`
	ModifiedAt  time.Time   `json:"modifiedAt"`
}

func (o *Order) PackItem(i item.Item, q int) {
	o.PackedItems = append(o.PackedItems, item.Item{Name: i.Name, Quantity: q})
}

const (
	Requested  = OrderStatus("REQUESTED")
	InProgress = OrderStatus("IN_PROGRESS")
	Ready      = OrderStatus("READY")
	Done       = OrderStatus("DONE")
)

type OrderStatus string
