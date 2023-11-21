package stack

import (
	"context"
	"fmt"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"log"
	"math/rand"
	"mc-burger-orders/event"
	"mc-burger-orders/testing/utils"
	utils2 "mc-burger-orders/utils"
	"testing"
)

var (
	stackTopic       = fmt.Sprintf("test-stack-events-%d", rand.Intn(100))
	stackKafkaConfig *event.TopicConfigs
)

func TestStack_Add(t *testing.T) {
	ctx := context.Background()
	kafkaContainer, brokers := utils.TestWithKafka(ctx)
	stackKafkaConfig = &event.TopicConfigs{
		Topic:             stackTopic,
		Brokers:           brokers,
		NumPartitions:     1,
		ReplicationFactor: 1,
		AutoCreateTopic:   true,
	}

	t.Run("should emit event when add one item", shouldEmitAddItemEvenWhenOneItemAdded)
	t.Run("should emit event when adding multiple items", shouldEmitAddItemEvenWhenMultipleItemsAdded)

	t.Cleanup(func() {
		log.Println("Running Clean UP code")
		utils.TerminateKafka(kafkaContainer)
	})
}

func shouldEmitAddItemEvenWhenOneItemAdded(t *testing.T) {
	// given
	events := make(chan kafka.Message)
	sut := NewEmptyStack()
	sut.ConfigureWriter(event.NewTopicWriter(stackKafkaConfig))

	reader := event.NewTopicReader(stackKafkaConfig, nil)
	go reader.SubscribeToTopic(events)

	// when
	sut.Add("hamburger")

	// then
	select {
	case newEvent := <-events:
		eventType, err := utils2.GetEventType(newEvent)
		if err != nil {
			assert.Fail(t, err.Error())
		}

		assert.Equal(t, ItemAddedToStackEvent, eventType)

		assert.Equal(t, "[{\"itemName\":\"hamburger\",\"quantity\":1}]", string(newEvent.Value))
		break
	}
}

func shouldEmitAddItemEvenWhenMultipleItemsAdded(t *testing.T) {
	// given
	events := make(chan kafka.Message)
	sut := NewEmptyStack()
	sut.ConfigureWriter(event.NewTopicWriter(stackKafkaConfig))

	reader := event.NewTopicReader(stackKafkaConfig, nil)
	go reader.SubscribeToTopic(events)

	// when
	sut.AddMany("cheeseburger", 10)

	// then
	select {
	case newEvent := <-events:
		eventType, err := utils2.GetEventType(newEvent)
		if err != nil {
			assert.Fail(t, err.Error())
		}

		assert.Equal(t, ItemAddedToStackEvent, eventType)

		assert.Equal(t, "[{\"itemName\":\"cheeseburger\",\"quantity\":10}]", string(newEvent.Value))
		break
	}
}
