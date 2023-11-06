package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"mc-burger-orders/item"
	"time"
)

type Order struct {
	Id          *primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	OrderNumber int64               `json:"orderNumber" bson:"orderNumber"`
	CustomerId  int                 `json:"customerId" bson:"customerId"`
	Items       []item.Item         `json:"items" bson:"items"`
	PackedItems []item.Item         `json:"packedItems" bson:"packedItems"`
	Status      OrderStatus         `json:"status" bson:"status" bson:"status"`
	CreatedAt   time.Time           `json:"createdAt" bson:"createdAt"`
	ModifiedAt  time.Time           `json:"modifiedAt" bson:"modifiedAt"`
}

func CreateNewOrder(number int64, order NewOrder) Order {
	objectID := primitive.NewObjectID()
	return Order{Id: &objectID, OrderNumber: number, CustomerId: order.CustomerId, Items: order.Items, Status: Requested, CreatedAt: time.Now(), ModifiedAt: time.Now()}
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
