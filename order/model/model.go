package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"mc-burger-orders/kitchen/item"
	"mc-burger-orders/log"
	"time"
)

type Order struct {
	Id          *primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	OrderNumber int64               `json:"orderNumber" bson:"orderNumber"`
	CustomerId  int                 `json:"customerId" bson:"customerId"`
	Items       []item.Item         `json:"items" bson:"items"`
	PackedItems []item.Item         `json:"packedItems" bson:"packedItems"`
	Status      OrderStatus         `json:"status" bson:"status"`
	CreatedAt   time.Time           `json:"createdAt" bson:"createdAt"`
	ModifiedAt  time.Time           `json:"modifiedAt" bson:"modifiedAt"`
}

func CreateNewOrder(number int64, order NewOrder) Order {
	objectID := primitive.NewObjectID()
	return Order{Id: &objectID, OrderNumber: number, CustomerId: order.CustomerId, Items: order.Items, Status: Requested, CreatedAt: time.Now(), ModifiedAt: time.Now()}
}

func (o *Order) PackItem(name string, quantity int) bool {
	if quantity > 0 {
		if o.PackedItems == nil {
			o.PackedItems = make([]item.Item, 0)
		}
		o.PackedItems = append(o.PackedItems, item.Item{Name: name, Quantity: quantity})
	}

	packedItemsCount := o.getItemsCount(o.PackedItems)
	if packedItemsCount == 0 {
		return false
	}

	var newStatus OrderStatus
	switch {
	case packedItemsCount < o.getItemsCount(o.Items):
		newStatus = InProgress
	case packedItemsCount == o.getItemsCount(o.Items):
		newStatus = Ready
	default:
		newStatus = Requested
	}
	if newStatus != o.Status {
		o.Status = newStatus
		// Push Order Status Updated
		log.Warning.Println("Order Status updated to:", o.Status)
		return true
	}
	return false
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
	COLLECTED  = OrderStatus("COLLECTED")
)

type OrderStatus string
