package utils

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"testing"
	"time"
)

func TestWithMongo(t *testing.T, ctx context.Context) (*mongodb.MongoDBContainer, *mongo.Database) {
	mongoDbContainer := startMongoContainer(t, ctx)
	mongoDb := getDatabase(t, mongoDbContainer, ctx)
	return mongoDbContainer, mongoDb
}

func TerminateMongo(t *testing.T, mongodbContainer *mongodb.MongoDBContainer) {
	ctx := context.Background()

	t.Log("Terminating MongoDB....")

	if err := mongodbContainer.Terminate(ctx); err != nil {
		assert.Failf(t, "Error while terminating Mongo Container", err.Error())
	}
}

func DeleteMany(t *testing.T, c *mongo.Collection, filter interface{}) {
	_, err := c.DeleteMany(context.TODO(), filter)

	if err != nil {
		assert.Failf(t, "failed to remove records after the test: %s", err.Error())
	}
}

func InsertMany(t *testing.T, c *mongo.Collection, records []interface{}) {
	_, err := c.InsertMany(context.TODO(), records)
	if err != nil {
		assert.Failf(t, "failed to insert test orders before the test: %s", err.Error())
	}

}

func startMongoContainer(t *testing.T, ctx context.Context) *mongodb.MongoDBContainer {
	mongodbContainer, err := mongodb.RunContainer(ctx, testcontainers.WithImage("mongo:6"))
	attempt := 1
	for {
		if err != nil && attempt < 4 {
			t.Log("MONGO SETUP: RunContainer attempt", attempt)

			time.Sleep(1 * time.Second)
			attempt++
			mongodbContainer, err = mongodb.RunContainer(ctx, testcontainers.WithImage("mongo:6"))
		}
		if err != nil && attempt == 4 {
			assert.Failf(t, "MONGO SETUP: Unable to Start MongoDB. %v", err.Error())
		} else {
			t.Log("MONGO SETUP: No error on attempt", attempt)
			break
		}
	}
	return mongodbContainer
}

func getDatabase(t *testing.T, m *mongodb.MongoDBContainer, ctx context.Context) *mongo.Database {
	endpoint, err := m.ConnectionString(ctx)
	if err != nil {
		assert.Failf(t, "MONGO SETUP: Unable to get MongoDB Connection String", err.Error())
	}

	t.Log("✅✅✅ Mongo Container: ", endpoint, ": ✅✅✅")
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(endpoint))
	if err != nil {
		assert.Failf(t, "MONGO SETUP: Unable to create Mongo client from Connection String: %v. %v", endpoint, err)
	}

	return mongoClient.Database("test-orders-db")
}
