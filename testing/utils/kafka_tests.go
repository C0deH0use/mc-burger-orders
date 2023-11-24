package utils

import (
	"context"
	kk "github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/kafka"
	"testing"
	"time"
)

func TestWithKafka(t *testing.T, ctx context.Context) (*kafka.KafkaContainer, []string) {
	kafkaContainer := startKafkaContainer(t, ctx)
	brokers, err := kafkaContainer.Brokers(ctx)
	if err != nil {
		assert.Failf(t, "KAFKA SETUP: Unable to connect to Cluster and read brokers. %v", err.Error())
	}

	t.Log("✅✅✅ Kafka Container is Up.... Brokers: ", brokers, ": ✅✅✅")
	return kafkaContainer, brokers
}

func startKafkaContainer(t *testing.T, ctx context.Context) *kafka.KafkaContainer {
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
			t.Log("KAFKA SETUP: No error on attempt", attempt)
			break
		}
	}

	return kafkaContainer
}

func TerminateKafka(t *testing.T, kafkaContainer *kafka.KafkaContainer) {
	ctx := context.Background()

	t.Log("Terminating Kafka")
	if err := kafkaContainer.Terminate(ctx); err != nil {
		assert.Fail(t, "Error while terminating Kafka Container")
	}
}

func TerminateKafkaReader(t *testing.T, testReader *kk.Reader) {
	err := testReader.Close()
	if err != nil {
		assert.Fail(t, "KAFKA TEARDOWN | failure when closing test reader")
	}
}
