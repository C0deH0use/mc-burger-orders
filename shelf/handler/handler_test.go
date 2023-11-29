package handler

import (
	"context"
	"fmt"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"mc-burger-orders/command"
	"mc-burger-orders/event"
	"mc-burger-orders/schedule"
	"mc-burger-orders/shelf"
	"mc-burger-orders/testing/utils"
	"sync"
	"testing"
)

var (
	shelfHandlerKafkaConfig *event.TopicConfigs
	shelfHandlerTopic       = fmt.Sprintf("test-shelf-handler-%d", rand.Intn(100))
)

func TestHandler_ShelfMessages(t *testing.T) {
	utils.IntegrationTest(t)
	ctx := context.Background()

	kafkaContainer, brokers := utils.TestWithKafka(t, ctx)
	shelfHandlerKafkaConfig = event.TestTopicConfigs(shelfHandlerTopic, brokers...)

	t.Run("should request new items when handling check favorite messages", shouldRequestNewItemsWhenHandlingCheckFavoriteMessages)

	t.Cleanup(func() {
		t.Log("Running Clean UP code")
		utils.TerminateKafka(t, ctx, kafkaContainer)
	})

}

func shouldRequestNewItemsWhenHandlingCheckFavoriteMessages(t *testing.T) {
	// given
	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(1)
	msgChan := make(chan kafka.Message)

	eventBus := event.NewInternalEventBus()
	kitchenStubService := shelf.NewShelfStubServiceWithWG(waitGroup)

	sut := &Handler{
		Shelf:          selfWithItems(),
		KitchenService: kitchenStubService,
		defaultHandler: command.DefaultCommandHandler{},
	}

	sendMessage()

	eventBus.AddHandler(sut)
	reader := event.NewTopicReader(shelfHandlerKafkaConfig, eventBus)

	// when
	go reader.SubscribeToTopic(msgChan)

	// then
	waitGroup.Wait()

	assert.True(t, kitchenStubService.HaveBeenCalledWith(shelf.RequestMatchingFnc("hamburger", 5)))
}

func sendMessage() {
	testWriter := event.NewTopicWriter(shelfHandlerKafkaConfig)
	_ = testWriter.SendMessage(context.Background(), schedule.CheckFavoritesOnShelfMessage())
}

func selfWithItems() *shelf.Shelf {
	s := shelf.NewEmptyShelf()
	s.AddMany("cheeseburger", 5)
	s.AddMany("spicy-stripes", 5)
	s.AddMany("hot-wings", 5)
	s.AddMany("fries", 5)
	return s
}
