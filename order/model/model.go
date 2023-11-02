package model

import (
	"mc-burger-orders/item"
	"time"
)

type Order struct {
	Id          int         `json:"_id"`
	OrderNumber int         `json:"orderNumber"`
	CustomerId  int         `json:"customerId"`
	Items       []item.Item `json:"items"`
	PackedItems []item.Item `json:"packedItems"`
	Status      OrderStatus `json:"status" bson:"status"`
	CreatedAt   time.Time   `json:"createdAt"`
	ModifiedAt  time.Time   `json:"modifiedAt"`
}

func (o *Order) PackItem(name string, quantity int) {
	o.PackedItems = append(o.PackedItems, item.Item{Name: name, Quantity: quantity})

	packedItemsCount := o.getItemsCount(o.PackedItems)
	if packedItemsCount == 0 {
		return
	}

	if packedItemsCount < o.getItemsCount(o.Items) {
		o.Status = InProgress
	} else {
		o.Status = Ready
	}
}

func (o *Order) getItemsCount(items []item.Item) int {
	var count = 0

	for _, i := range items {
		count += i.Quantity
	}

	return count
}

const (
	Requested  = OrderStatus("REQUESTED")
	InProgress = OrderStatus("IN_PROGRESS")
	Ready      = OrderStatus("READY")
	Done       = OrderStatus("DONE")
)

type OrderStatus string
