package order

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"mc-burger-orders/log"
	"time"
)

type FetchByIdRepository interface {
	FetchById(ctx context.Context, id interface{}) (*Order, error)
}

type FetchByMissingItemsOrderByOrderNumber interface {
	FetchByMissingItem(ctx context.Context, itemName string) ([]*Order, error)
}

type FetchByOrderNumberRepository interface {
	FetchByOrderNumber(ctx context.Context, orderNumber int64) (*Order, error)
}

type FetchManyRepository interface {
	FetchMany(ctx context.Context) ([]*Order, error)
}

type StoreRepository interface {
	InsertOrUpdate(ctx context.Context, order *Order) (*Order, error)
}

type OrderRepository interface {
	StoreRepository
	FetchByIdRepository
	FetchByOrderNumberRepository
	FetchByMissingItemsOrderByOrderNumber
	FetchManyRepository
}

type PackingOrderItemsRepository interface {
	StoreRepository
	FetchByMissingItemsOrderByOrderNumber
}

type OrderRepositoryImpl struct {
	c *mongo.Collection
}

func NewRepository(database *mongo.Database) *OrderRepositoryImpl {
	collection := database.Collection("orders")
	return &OrderRepositoryImpl{c: collection}
}

func (r *OrderRepositoryImpl) InsertOrUpdate(ctx context.Context, order *Order) (*Order, error) {
	order.ModifiedAt = time.Now()
	log.Info.Printf("Updating existing Order Number: %v", order.OrderNumber)
	filterDef := bson.D{{Key: "orderNumber", Value: order.OrderNumber}}
	updateDef := bson.D{{Key: "$set", Value: order}}
	upsertOption := true
	updateOptions := &options.UpdateOptions{
		Upsert: &upsertOption,
	}

	result, err := r.c.UpdateOne(ctx, filterDef, updateDef, updateOptions)
	if err != nil {
		log.Error.Println("Error when updating order in db", err)
		return nil, err
	}
	if order.Id == nil && result.UpsertedID == nil {
		err = fmt.Errorf("missing order pk id for the document")
		return nil, err
	}

	var orderId primitive.ObjectID
	switch {
	case result.UpsertedID != nil:
		orderId = result.UpsertedID.(primitive.ObjectID)
	case order.Id != nil:
		orderId = *order.Id
	default:
		err = fmt.Errorf("cannot determin order pk id")
		return nil, err
	}

	return r.FetchById(ctx, orderId)
}

func (r *OrderRepositoryImpl) FetchById(ctx context.Context, id interface{}) (*Order, error) {
	filter := bson.D{{Key: "_id", Value: id}}
	result := r.c.FindOne(ctx, filter)
	if result.Err() != nil {
		log.Error.Println("Error when fetching order by id", id, result.Err())
		return nil, result.Err()
	}

	var order *Order
	if err := result.Decode(&order); err != nil {
		log.Error.Println("Error reading Order raw data", err)
		return nil, err
	}

	return order, nil
}

func (r *OrderRepositoryImpl) FetchByOrderNumber(ctx context.Context, orderNumber int64) (*Order, error) {
	filter := bson.D{{Key: "orderNumber", Value: orderNumber}}
	result := r.c.FindOne(ctx, filter)
	if result.Err() != nil {
		log.Error.Println("Error when fetching order by orderNumber", orderNumber, result.Err())
		return nil, result.Err()
	}

	var order *Order
	if err := result.Decode(&order); err != nil {
		log.Error.Println("Error reading Order raw data", err)
		return nil, err
	}

	return order, nil
}

func (r *OrderRepositoryImpl) FetchMany(ctx context.Context) ([]*Order, error) {
	filterDef := bson.D{
		{
			Key: "status",
			Value: bson.D{{
				Key:   "$in",
				Value: bson.A{Requested, InProgress, Ready},
			}},
		},
	}
	findOptions := &options.FindOptions{
		Sort: bson.D{{
			Key:   "orderNumber",
			Value: 1,
		}},
	}
	cursor, err := r.c.Find(ctx, filterDef, findOptions)
	if err != nil {
		log.Error.Println("Error when fetching orders from db", err)
		return nil, err
	}

	var orders []*Order
	if err = cursor.All(ctx, &orders); err != nil {
		log.Error.Println("Error reading cursor data", err)
		return nil, err
	}

	return orders, nil
}

func (r *OrderRepositoryImpl) FetchByMissingItem(ctx context.Context, itemName string) ([]*Order, error) {
	filterDef := bson.D{
		{
			Key:   "items.name",
			Value: itemName,
		},
		{
			Key: "status",
			Value: bson.D{{
				Key:   "$in",
				Value: bson.A{Requested, InProgress},
			}},
		},
	}
	findOptions := &options.FindOptions{
		Sort: bson.D{{
			Key:   "orderNumber",
			Value: 1,
		}},
	}
	cursor, err := r.c.Find(ctx, filterDef, findOptions)
	if err != nil {
		log.Error.Println("Error when fetching dbRecords from db", err)
		return nil, err
	}

	dbRecords := make([]*Order, 0)
	if err = cursor.All(ctx, &dbRecords); err != nil {
		log.Error.Println("Error reading cursor data", err)
		return nil, err
	}
	orders := make([]*Order, 0)
	for _, record := range dbRecords {
		if count, err := record.GetMissingItemsCount(itemName); err == nil && count > 0 {
			orders = append(orders, record)
		}
	}

	return orders, nil
}
