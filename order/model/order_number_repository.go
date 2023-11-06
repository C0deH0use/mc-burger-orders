package model

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

func NewOrderNumberRepository(database *mongo.Database) *OrderNumberRepositoryImpl {
	collection := database.Collection("order-numbers")
	return &OrderNumberRepositoryImpl{c: collection}
}

type OrderNumberRepository interface {
	GetNext(ctx context.Context) (int64, error)
}

type OrderNumberRepositoryImpl struct {
	c *mongo.Collection
}

type FetchNextOrderNumberRepository interface {
	GetNext(ctx context.Context) (int, error)
}

func (r *OrderNumberRepositoryImpl) GetNext(ctx context.Context) (int64, error) {
	fmt.Println("Get next order number")
	limit := int64(1)
	sortDef := map[string]int{
		"number": 1,
	}

	opts := &options.FindOptions{
		Limit: &limit,
		Sort:  sortDef,
	}
	cursor, err := r.c.Find(ctx, bson.D{}, opts)
	if err != nil {
		log.Println("Error when fetching current latest order number from db", err)
		return -1, err
	}
	var numbers []OrderNumber
	if err = cursor.All(ctx, &numbers); err != nil {
		log.Println("Error reading cursor data", err)
		return -1, err
	}
	nextOrderNumber, err := r.getAndPersistNext(ctx, numbers)
	if err != nil {
		log.Println("Error Determining the next order number", err)
		return -1, err
	}

	log.Println("Next Order Number is", nextOrderNumber)
	return nextOrderNumber, nil
}

func (r *OrderNumberRepositoryImpl) getAndPersistNext(ctx context.Context, numbers []OrderNumber) (int64, error) {
	var latestNumber = NewOrderNumber(1)
	if len(numbers) > 0 {
		lastNumber := numbers[0].Number
		log.Println("Last Order number", latestNumber)

		latestNumber = NewOrderNumber(lastNumber + 1)
	}

	_, err := r.c.InsertOne(ctx, latestNumber)
	if err != nil {
		log.Println("Error Persisting next order number", err)
		return -1, err
	}
	log.Println("Next order number persisted")

	return latestNumber.Number, nil
}