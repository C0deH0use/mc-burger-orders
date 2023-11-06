package utils

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
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

	log.Print("✅✅✅ Mongo Container: ", endpoint, ": ✅✅✅")
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
		deleteErr := fmt.Errorf("failed to remove records after the test: %s", err)
		panic(deleteErr)
	}
}

func InsertMany(c *mongo.Collection, records []interface{}) {
	_, err := c.InsertMany(context.TODO(), records)
	if err != nil {
		insertErr := fmt.Errorf("failed to insert test orders before the test: %s", err)
		panic(insertErr)
	}

}
