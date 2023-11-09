package model

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	InsertOrUpdate(ctx context.Context, order Order) (*Order, error)
}

type OrderRepository interface {
	StoreRepository
	FetchByIdRepository
	FetchManyRepository
}

type OrderRepositoryImpl struct {
	c *mongo.Collection
}

func NewRepository(database *mongo.Database) *OrderRepositoryImpl {
	collection := database.Collection("orders")
	return &OrderRepositoryImpl{c: collection}
}

func (r *OrderRepositoryImpl) InsertOrUpdate(ctx context.Context, order Order) (*Order, error) {
	order.ModifiedAt = time.Now()
	log.Println("Updating existing Order => ", order)
	filterDef := bson.D{{"orderNumber", order.OrderNumber}}
	updateDef := bson.D{{"$set", order}}
	upsertOption := true
	updateOptions := &options.UpdateOptions{
		Upsert: &upsertOption,
	}

	result, err := r.c.UpdateOne(ctx, filterDef, updateDef, updateOptions)
	if err != nil {
		log.Println("Error when updating order in db", err)
		return nil, err
	}

	return r.FetchById(ctx, result.UpsertedID)
}

func (r *OrderRepositoryImpl) FetchById(ctx context.Context, id interface{}) (*Order, error) {
	filter := bson.D{{"_id", id}}
	result := r.c.FindOne(ctx, filter)
	if result.Err() != nil {
		log.Println("Error when fetching order by id", id, result.Err())
		return nil, result.Err()
	}

	var order *Order
	if err := result.Decode(&order); err != nil {
		log.Println("Error reading Order raw data", err)
		return nil, err
	}

	return order, nil
}

func (r *OrderRepositoryImpl) FetchByOrderNumber(ctx context.Context, orderNumber int64) (*Order, error) {
	filter := bson.D{{"orderNumber", orderNumber}}
	result := r.c.FindOne(ctx, filter)
	if result.Err() != nil {
		log.Println("Error when fetching order by orderNumber", orderNumber, result.Err())
		return nil, result.Err()
	}

	var order *Order
	if err := result.Decode(&order); err != nil {
		log.Println("Error reading Order raw data", err)
		return nil, err
	}

	return order, nil
}

func (r *OrderRepositoryImpl) FetchMany(ctx context.Context) ([]Order, error) {
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
