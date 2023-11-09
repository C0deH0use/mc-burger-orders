package utils

import (
	"context"
	"fmt"
	kafkago "github.com/segmentio/kafka-go"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/kafka"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

func TestWithMongo(ctx context.Context) *mongodb.MongoDBContainer {
	mongodbContainer, err := mongodb.RunContainer(ctx,
		testcontainers.WithImage("mongo:6"),
	)
	if err != nil {
		panic(err)
	}

	return mongodbContainer
}

func TestWithKafka(ctx context.Context) *kafka.KafkaContainer {
	kafkaContainer, err := kafka.RunContainer(ctx, testcontainers.WithImage("confluentinc/confluent-local:7.5.0"))

	if err != nil {
		panic(err)
	}

	return kafkaContainer
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

func TerminateKafka(kafkaContainer *kafka.KafkaContainer) {
	ctx := context.Background()

	if err := kafkaContainer.Terminate(ctx); err != nil {
		panic(err)
	}
}

func TerminateKafkaReader(testReader *kafkago.Reader) {
	err := testReader.Close()
	if err != nil {
		panic("failure when closing test reader")
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
