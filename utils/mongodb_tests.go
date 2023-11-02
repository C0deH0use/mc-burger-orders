package utils

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
)

func TestWithMongo(t *testing.T) *mongodb.MongoDBContainer {
	ctx := context.Background()
	mongodbContainer, err := mongodb.RunContainer(ctx,
		testcontainers.WithImage("mongo:6"),
	)
	if err != nil {
		panic(err)
	}

	return mongodbContainer
}

func GetMongoDbFrom(m *mongodb.MongoDBContainer) *mongo.Database {
	ctx := context.Background()
	endpoint, err := m.ConnectionString(ctx)
	if err != nil {
		panic(err)
	}

	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(endpoint))
	if err != nil {
		panic(err)
	}

	return mongoClient.Database("test-orders-db")
}

func TerminateMongo(mongodbContainer *mongodb.MongoDBContainer) {
	ctx := context.Background()

	if err := mongodbContainer.Terminate(ctx); err != nil {
		panic(err)
	}
}

func DeleteMany(c *mongo.Collection, filter interface{}) {
	_, err := c.DeleteMany(context.TODO(), filter)

	if err != nil {
		panic("Failed to remove records after the test")
	}
}

func InsertMany(c *mongo.Collection, records []interface{}) {
	_, err := c.InsertMany(context.TODO(), records)
	if err != nil {
		panic("Failed to insert test orders before the test")
	}

}
