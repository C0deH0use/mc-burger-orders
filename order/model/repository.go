package model

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"time"
)

type FetchByIdRepository interface {
	FetchById(ctx context.Context, id interface{}) (*Order, error)
}

type FetchManyRepository interface {
	FetchMany(ctx context.Context) ([]Order, error)
}

type StoreRepository interface {
	Create(ctx context.Context, newOrder NewOrder) (*Order, error)
}

type OrderRepository struct {
	c *mongo.Collection
}

func NewRepository(database *mongo.Database) *OrderRepository {
	collection := database.Collection("orders")
	return &OrderRepository{c: collection}
}

func (r *OrderRepository) Create(ctx context.Context, no NewOrder) (*Order, error) {

	fmt.Println("Storing new Order request => ", no)
	orderNumber := 1010 // nextOrderNumber()
	o := &Order{OrderNumber: orderNumber, CustomerId: no.CustomerId, Items: no.Items, Status: Requested, CreatedAt: time.Now(), ModifiedAt: time.Now()}

	result, err := r.c.InsertOne(ctx, o)
	if err != nil {
		log.Println("Error when fetching orders from db", err)
		return nil, err
	}

	return r.FetchById(ctx, result.InsertedID)
}

func (r *OrderRepository) FetchById(ctx context.Context, id interface{}) (*Order, error) {
	filter := bson.D{{"_id", id}}
	result := r.c.FindOne(ctx, filter)
	if result.Err() != nil {
		log.Println("Error when fetching order by id", id, result.Err())
		return nil, result.Err()
	}

	var order *Order
	if err := result.Decode(order); err != nil {
		log.Println("Error reading Order raw data", err)
		return nil, err
	}

	return order, nil
}

func (r *OrderRepository) FetchMany(ctx context.Context) ([]Order, error) {
	cursor, err := r.c.Find(ctx, bson.D{})
	if err != nil {
		log.Println("Error when fetching orders from db", err)
		return nil, err
	}

	var orders []Order
	if err = cursor.All(context.TODO(), &orders); err != nil {
		log.Println("Error reading cursor data", err)
		return nil, err
	}

	return orders, nil
}
