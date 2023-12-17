package order

import (
	"fmt"
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

func CreateNewOrder(number int64, order NewOrder) *Order {
	objectID := primitive.NewObjectID()
	return &Order{Id: &objectID, OrderNumber: number, CustomerId: order.CustomerId, Items: order.Items, Status: Requested, CreatedAt: time.Now(), ModifiedAt: time.Now()}
}

func (o *Order) PackItem(name string, quantity int) bool {
	if quantity > 0 {
		if o.PackedItems == nil {
			o.PackedItems = make([]item.Item, 0)
		}
		o.PackedItems = append(o.PackedItems, item.Item{Name: name, Quantity: quantity})
	}

	packedItemsCount := o.GetItemsCount(o.PackedItems)
	if packedItemsCount == 0 {
		return false
	}

	var newStatus OrderStatus
	switch {
	case packedItemsCount < o.GetItemsCount(o.Items):
		newStatus = InProgress
	case packedItemsCount == o.GetItemsCount(o.Items):
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

func (o *Order) GetItemsCount(items []item.Item) int {
	var count = 0

	for _, i := range items {
		count += i.Quantity
	}

	return count
}

func (o *Order) GetMissingItems() []item.Item {
	i := make([]item.Item, 0)

	for _, ii := range o.Items {
		missingItemsCount, err := o.GetMissingItemsCount(ii.Name)
		if err != nil {
			log.Error.Printf("Order %d has incorrect items configuration, item: `%v` => %v", o.OrderNumber, ii.Name, err.Error())
			continue
		}
		if missingItemsCount > 0 {
			i = append(i, item.Item{Name: ii.Name, Quantity: missingItemsCount})
		}
	}
	return i
}

func (o *Order) GetMissingItemsCount(itemName string) (int, error) {
	var quantity = -1
	for _, i := range o.Items {
		if i.Name == itemName {
			quantity = i.Quantity
		}
	}
	if quantity == -1 {
		err := fmt.Errorf("could not find item `%v` on the order list", itemName)
		return -quantity, err
	}

	for _, i := range o.PackedItems {
		if i.Name == itemName {
			quantity -= i.Quantity
		}
	}

	return quantity, nil
}

const (
	Requested  = OrderStatus("REQUESTED")
	InProgress = OrderStatus("IN_PROGRESS")
	Ready      = OrderStatus("READY")
	Collected  = OrderStatus("COLLECTED")
)

type OrderStatus string
