package model

import (
	"mc-burger-orders/item"
	"time"
)

type Order struct {
	ID          int         `json:"id"`
	CustomerId  int         `json:"customerId"`
	Items       []item.Item `json:"items"`
	PackedItems []item.Item `json:"packedItems"`
	Status      OrderStatus `json:"status"`
	CreatedAt   time.Time   `json:"createdAt"`
	ModifiedAt  time.Time   `json:"modifiedAt"`
}

func (o *Order) PackItem(i item.Item, q int) {
	o.PackedItems = append(o.PackedItems, item.Item{Name: i.Name, Quantity: q})
}

const (
	Requested  = _status("REQUESTED")
	InProgress = _status("IN_PROGRESS")
	Ready      = _status("READY")
	Done       = _status("DONE")
)

type OrderStatus interface {
	_isStatus()
	Value() string
}

type _status string

func (_status) _isStatus() {}

func (_c _status) Value() string {
	return string(_c)
}
