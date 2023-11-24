package utils

import (
	"context"
	kafkago "github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/kafka"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"testing"
	"time"
)

func TestWithMongo(t *testing.T, ctx context.Context) *mongodb.MongoDBContainer {
	return tryRunMongoContainer(t, ctx)
}

func TestWithKafka(t *testing.T, ctx context.Context) (*kafka.KafkaContainer, []string) {
	kafkaContainer := tryRunKafkaContainer(t, ctx)
	brokers, err := kafkaContainer.Brokers(ctx)
	if err != nil {
		assert.Failf(t, "KAFKA SETUP: Unable to connect to Cluster and read brokers. %v", err.Error())
	}

	t.Log("✅✅✅ Kafka Container is Up.... Brokers: ", brokers, ": ✅✅✅")
	return kafkaContainer, brokers
}

func tryRunMongoContainer(t *testing.T, ctx context.Context) *mongodb.MongoDBContainer {
	mongodbContainer, err := mongodb.RunContainer(ctx, testcontainers.WithImage("mongo:6"))
	attempt := 1
	for {
		if err != nil && attempt < 4 {
			time.Sleep(1 * time.Second)
			attempt++
			mongodbContainer, err = mongodb.RunContainer(ctx, testcontainers.WithImage("mongo:6"))
		}
		if err != nil && attempt == 4 {
			assert.Failf(t, "MONGO SETUP: Unable to Start MongoDB. %v", err.Error())
		} else {
			break
		}
	}

	return mongodbContainer
}

func tryRunKafkaContainer(t *testing.T, ctx context.Context) *kafka.KafkaContainer {
	kafkaContainer, err := kafka.RunContainer(ctx, testcontainers.WithImage("confluentinc/confluent-local:7.5.0"))

	attempt := 1
	for {
		if err != nil && attempt < 4 {
			t.Log("KAFKA SETUP: RunContainer attempt", attempt)
			time.Sleep(1 * time.Second)
			attempt++
			kafkaContainer, err = kafka.RunContainer(ctx, testcontainers.WithImage("confluentinc/confluent-local:7.5.0"))
		}
		if err != nil && attempt == 4 {
			assert.Failf(t, "KAFKA SETUP: Unable to Start Kafka cluster. %v", err.Error())
		} else {
			break
		}
	}

	return kafkaContainer
}

func GetMongoDbFrom(t *testing.T, m *mongodb.MongoDBContainer) *mongo.Database {
	ctx := context.Background()
	endpoint, err := m.ConnectionString(ctx)
	if err != nil {
		assert.Failf(t, "MONGO SETUP: Unable to get MongoDB Connection String. %v", err.Error())
	}

	log.Print("✅✅✅ Mongo Container: ", endpoint, ": ✅✅✅")
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(endpoint))
	if err != nil {
		assert.Failf(t, "MONGO SETUP: Unable to create Mongo client from Connection String: %v. %v", endpoint, err)
	}

	return mongoClient.Database("test-orders-db")
}

func TerminateMongo(t *testing.T, mongodbContainer *mongodb.MongoDBContainer) {
	ctx := context.Background()

	t.Log("Terminating MongoDB....")

	if err := mongodbContainer.Terminate(ctx); err != nil {
		assert.Failf(t, "Error while terminating Mongo Container", err.Error())
	}
}

func TerminateKafka(t *testing.T, kafkaContainer *kafka.KafkaContainer) {
	ctx := context.Background()

	t.Log("Terminating Kafka")
	if err := kafkaContainer.Terminate(ctx); err != nil {
		assert.Fail(t, "Error while terminating Kafka Container")
	}
}

func TerminateKafkaReader(t *testing.T, testReader *kafkago.Reader) {
	err := testReader.Close()
	if err != nil {
		assert.Fail(t, "KAFKA TEARDOWN | failure when closing test reader")
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
