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
	"time"
)

func TestWithMongo(ctx context.Context) *mongodb.MongoDBContainer {
	return tryRunMongoContainer(ctx)
}

func TestWithKafka(ctx context.Context) (*kafka.KafkaContainer, []string) {
	kafkaContainer := tryRunKafkaContainer(ctx)
	brokers, err := kafkaContainer.Brokers(ctx)
	if err != nil {
		log.Panicf("KAFKA SETUP: Unable to connect to Cluster and read brokers. %v", err)
	}

	log.Print("✅✅✅ Kafka Container is Up.... Brokers: ", brokers, ": ✅✅✅")
	return kafkaContainer, brokers
}

func tryRunMongoContainer(ctx context.Context) *mongodb.MongoDBContainer {
	mongodbContainer, err := mongodb.RunContainer(ctx, testcontainers.WithImage("mongo:6"))
	attempt := 1
	for {
		if err != nil && attempt < 4 {
			time.Sleep(250 * time.Millisecond)
			attempt++
			if err = mongodbContainer.Terminate(ctx); err != nil {
				log.Panicf("MongoDB SETUP: Unable to Terminate MongoDB when trying to RunContainer. %v", err)
			}
			mongodbContainer, err = mongodb.RunContainer(ctx, testcontainers.WithImage("mongo:6"))
		}
		if err != nil && attempt == 4 {
			log.Panicf("MONGO SETUP: Unable to Start MongoDB. %v", err)
		} else {
			break
		}
	}

	return mongodbContainer
}

func tryRunKafkaContainer(ctx context.Context) *kafka.KafkaContainer {
	kafkaContainer, err := kafka.RunContainer(ctx, testcontainers.WithImage("confluentinc/confluent-local:7.5.0"))

	attempt := 1
	for {
		if err != nil && attempt < 4 {
			time.Sleep(250 * time.Millisecond)
			attempt++
			if err = kafkaContainer.Terminate(ctx); err != nil {
				log.Panicf("KAFKA SETUP: Unable to Terminate Kafka cluster when trying to RunContainer. %v", err)
			}
			kafkaContainer, err = kafka.RunContainer(ctx, testcontainers.WithImage("confluentinc/confluent-local:7.5.0"))
		}
		if err != nil && attempt == 4 {
			log.Panicf("KAFKA SETUP: Unable to Start Kafka cluster. %v", err)
		} else {
			break
		}
	}

	return kafkaContainer
}

func GetMongoDbFrom(m *mongodb.MongoDBContainer) *mongo.Database {
	ctx := context.Background()
	endpoint, err := m.ConnectionString(ctx)
	if err != nil {
		log.Panicf("MONGO SETUP: Unable to get MongoDB Connection String. %v", err)
	}

	log.Print("✅✅✅ Mongo Container: ", endpoint, ": ✅✅✅")
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(endpoint))
	if err != nil {
		log.Panicf("MONGO SETUP: Unable to create Mongo client from Connection String: %v. %v", endpoint, err)
	}

	return mongoClient.Database("test-orders-db")
}

func TerminateMongo(mongodbContainer *mongodb.MongoDBContainer) {
	ctx := context.Background()

	log.Print("Terminating MongoDB....")

	if err := mongodbContainer.Terminate(ctx); err != nil {
		log.Println("Error while terminating Mongo Container", err)
	}
}

func TerminateKafka(kafkaContainer *kafka.KafkaContainer) {
	ctx := context.Background()

	log.Print("Terminating Kafka")
	if err := kafkaContainer.Terminate(ctx); err != nil {
		log.Println("Error while terminating Kafka Container", err)
	}
}

func TerminateKafkaReader(testReader *kafkago.Reader) {
	err := testReader.Close()
	if err != nil {
		log.Panicf("KAFKA TEARDOWN | failure when closing test reader")
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
