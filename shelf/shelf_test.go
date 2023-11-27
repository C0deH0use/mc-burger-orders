package shelf

import (
	"context"
	"fmt"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	kafka2 "github.com/testcontainers/testcontainers-go/modules/kafka"
	"log"
	"math/rand"
	"mc-burger-orders/command"
	"mc-burger-orders/event"
	"mc-burger-orders/testing/utils"
	utils2 "mc-burger-orders/utils"
	"sync"
	"testing"
)

var (
	brokers          []string
	kafkaContainer   *kafka2.KafkaContainer
	stackKafkaConfig *event.TopicConfigs
)

type StubCommand struct {
	Message kafka.Message
	waitG   *sync.WaitGroup
}

func (s *StubCommand) Execute(ctx context.Context, message kafka.Message) (bool, error) {
	s.Message = message
	s.waitG.Done()
	return true, nil
}

func TestStack_Add(t *testing.T) {
	utils.IntegrationTest(t)
	ctx := context.Background()
	kafkaContainer, brokers = utils.TestWithKafka(t, ctx)

	log.Printf("Starting Shelf Test tests....")

	t.Run("should emit event when add one item", shouldEmitAddItemEvenWhenOneItemAdded)
	t.Run("should emit event when adding multiple items", shouldEmitAddItemEvenWhenMultipleItemsAdded)

	t.Cleanup(func() {
		t.Log("Running Clean UP code")
		utils.TerminateKafka(t, ctx, kafkaContainer)
	})
}

func shouldEmitAddItemEvenWhenOneItemAdded(t *testing.T) {
	// given
	events := make(chan kafka.Message)
	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(1)
	stubCommand := StubCommand{waitG: waitGroup}

	eventBus := event.NewInternalEventBus()
	commandHandler := command.NewCommandHandler()
	commandHandler.AddCommands(ItemAddedOnShelfEvent, &stubCommand)
	eventBus.AddHandler(commandHandler)

	topicName := fmt.Sprintf("test-stack-events-%d", rand.Intn(100))
	stackKafkaConfig = event.TestTopicConfigs(topicName, brokers...)

	sut := NewEmptyShelf()
	sut.ConfigureWriter(event.NewTopicWriter(stackKafkaConfig))

	reader := event.NewTopicReader(stackKafkaConfig, eventBus)
	go reader.SubscribeToTopic(events)

	// when
	sut.Add("hamburger")

	// then
	waitGroup.Wait()
	newEvent := stubCommand.Message
	if newEvent.Topic != topicName {
		assert.Fail(t, fmt.Sprintf("new event from topic %v, it IS NOT THE CORRECT topic", newEvent.Topic))
		return
	}
	log.Printf("new event from topic %v, it's the correct topic", newEvent.Topic)

	eventType, err := utils2.GetEventType(newEvent)
	if err != nil {
		assert.Fail(t, err.Error())
		return
	}

	assert.Equal(t, ItemAddedOnShelfEvent, eventType)
	assert.Equal(t, "[{\"itemName\":\"hamburger\",\"quantity\":1}]", string(newEvent.Value))
}

func shouldEmitAddItemEvenWhenMultipleItemsAdded(t *testing.T) {
	// given
	events := make(chan kafka.Message)
	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(1)
	stubCommand := StubCommand{waitG: waitGroup}

	eventBus := event.NewInternalEventBus()
	commandHandler := command.NewCommandHandler()
	commandHandler.AddCommands(ItemAddedOnShelfEvent, &stubCommand)
	eventBus.AddHandler(commandHandler)

	topicName := fmt.Sprintf("test-stack-events-%d", rand.Intn(100))
	stackKafkaConfig = event.TestTopicConfigs(topicName, brokers...)

	sut := NewEmptyShelf()
	sut.ConfigureWriter(event.NewTopicWriter(stackKafkaConfig))

	reader := event.NewTopicReader(stackKafkaConfig, eventBus)
	go reader.SubscribeToTopic(events)

	// when
	sut.AddMany("cheeseburger", 10)

	// then
	waitGroup.Wait()
	newEvent := stubCommand.Message
	if newEvent.Topic != topicName {
		assert.Fail(t, fmt.Sprintf("new event from topic %v, it IS NOT THE CORRECT topic", newEvent.Topic))
		return
	}

	log.Printf("new event from topic %v, it's the correct topic", newEvent.Topic)
	eventType, err := utils2.GetEventType(newEvent)
	log.Printf("||new event %v read. Body: %v %+v", eventType, string(newEvent.Value), newEvent)

	if err != nil {
		assert.Fail(t, err.Error())
		return
	}

	assert.Equal(t, ItemAddedOnShelfEvent, eventType)

	assert.Equal(t, "[{\"itemName\":\"cheeseburger\",\"quantity\":10}]", string(newEvent.Value))
}
