package model

import (
	"fmt"
	"mc-burger-orders/item"
	"time"
)

type OrderRepository struct {
}

func (r *OrderRepository) CreateOrder(no NewOrder) *Order {

	fmt.Println("Storing new Order request => ", no)
	// TODO Repository.call
	o := &Order{ID: 1010, CustomerId: no.CustomerId, Items: no.Items, Status: Requested, CreatedAt: time.Now(), ModifiedAt: time.Now()}

	//create();
	return o
}

func (r *OrderRepository) FetchOrders() []Order {
	// TODO Repository.call
	orders := []Order{
		{ID: 1000, CustomerId: 1, Items: []item.Item{{Name: "hamburger", Quantity: 1}, {Name: "fries", Quantity: 1}}, Status: Ready, CreatedAt: time.Now(), ModifiedAt: time.Now()},
		{ID: 1001, CustomerId: 2, Items: []item.Item{{Name: "hamburger", Quantity: 1}, {Name: "cheeseburger", Quantity: 2}}, Status: InProgress, CreatedAt: time.Now(), ModifiedAt: time.Now()},
		{ID: 1002, CustomerId: 3, Items: []item.Item{{Name: "cheeseburger", Quantity: 2}, {Name: "cheeseburger", Quantity: 3}}, Status: Requested, CreatedAt: time.Now(), ModifiedAt: time.Now()},
	}

	return orders
}
