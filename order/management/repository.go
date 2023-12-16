package management

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"mc-burger-orders/log"
	"mc-burger-orders/order"
	"time"
)

type OrderRepository interface {
	PackingOrderItemsRepository
}

type PackingOrderItemsRepository interface {
	FetchOrdersWithMissingPackedItems(ctx context.Context) ([]order.Order, error)
}

func NewOrderRepository(db *mongo.Database) OrderRepository {
	collection := db.Collection("orders")
	return &OrderRepositoryImpl{c: collection}
}

type OrderRepositoryImpl struct {
	c *mongo.Collection
}

func (s *OrderRepositoryImpl) FetchOrdersWithMissingPackedItems(ctx context.Context) ([]order.Order, error) {
	now := time.Now()
	filterDef := bson.D{
		{
			Key: "status",
			Value: bson.D{{
				Key:   "$in",
				Value: bson.A{order.Requested, order.InProgress},
			}},
		},
		{
			Key: "modifiedAt",
			Value: bson.D{{
				Key:   "$lte",
				Value: now.Add(time.Minute * -2),
			}},
		},
	}
	findOptions := &options.FindOptions{
		Sort: bson.D{{
			Key:   "orderNumber",
			Value: 1,
		}},
	}
	cursor, err := s.c.Find(ctx, filterDef, findOptions)
	if err != nil {
		log.Error.Println("Error when fetching dbRecords from db", err)
		return make([]order.Order, 0), err
	}

	dbRecords := make([]order.Order, 0)
	if err = cursor.All(ctx, &dbRecords); err != nil {
		log.Error.Println("Error reading cursor data", err)
		return dbRecords, err
	}
	return dbRecords, nil
}
