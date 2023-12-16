package management

import (
	"context"
	"fmt"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"math/rand"
	"mc-burger-orders/command"
	"mc-burger-orders/event"
	"mc-burger-orders/kitchen/item"
	"mc-burger-orders/order"
	"mc-burger-orders/testing/utils"
	"strconv"
	"sync"
	"testing"
	"time"
)

var (
	database             *mongo.Database
	collectionDb         *mongo.Collection
	mongoContainer       *mongodb.MongoDBContainer
	orderJobsKafkaConfig *event.TopicConfigs
	orderJobTopic        = fmt.Sprintf("test-order-jobs-events-%d", rand.Intn(100))
)

func TestOrdersHandler_Handle(t *testing.T) {
	utils.IntegrationTest(t)
	ctx := context.Background()

	mongoContainer, database = utils.TestWithMongo(t, ctx)
	kafkaContainer, brokers := utils.TestWithKafka(t, ctx)
	orderJobsKafkaConfig = event.TestTopicConfigs(orderJobTopic, brokers...)

	collectionDb = database.Collection("orders")

	t.Run("should pack item when shelf event occurred", shouldPackPreparedItemWhenEvenFromStackOccurred)

	t.Cleanup(func() {
		t.Log("Running Clean UP code")
		utils.TerminateMongo(t, ctx, mongoContainer)
		utils.TerminateKafka(t, ctx, kafkaContainer)
	})
}

func shouldPackPreparedItemWhenEvenFromStackOccurred(t *testing.T) {
	// given
	wg := &sync.WaitGroup{}
	wg.Add(3)
	eventBus := event.NewInternalEventBus()
	currentTime := time.Now()
	expectedOrders := []interface{}{
		order.Order{OrderNumber: 999, CustomerId: 10, Items: []item.Item{{Name: "hamburger", Quantity: 1}, {Name: "fries", Quantity: 1}}, Status: order.Ready, ModifiedAt: currentTime},
		order.Order{OrderNumber: 1000, CustomerId: 1, Items: []item.Item{{Name: "hamburger", Quantity: 1}, {Name: "fries", Quantity: 1}}, PackedItems: []item.Item{{Name: "fries", Quantity: 1}}, Status: order.Requested, ModifiedAt: currentTime.Add(time.Minute * -3)},
		order.Order{OrderNumber: 1002, CustomerId: 3, Items: []item.Item{{Name: "cheeseburger", Quantity: 2}, {Name: "hamburger", Quantity: 3}}, PackedItems: []item.Item{{Name: "hamburger", Quantity: 2}}, Status: order.InProgress, ModifiedAt: currentTime.Add(time.Minute * -10)},
	}
	utils.DeleteMany(t, collectionDb, bson.D{})
	utils.InsertMany(t, collectionDb, expectedOrders)

	queryService := NewOrderQueryService(NewOrderRepository(database))
	kitchenStubService := NewKitchenStubService(wg)

	handler := &OrderManagementHandler{
		queryService:   queryService,
		kitchenService: kitchenStubService,
		defaultHandler: command.DefaultCommandHandler{},
	}

	eventBus.AddHandler(handler)
	topicReader := event.NewTopicReader(orderJobsKafkaConfig, eventBus)
	go topicReader.SubscribeToTopic(make(chan kafka.Message))

	// when
	go sendMessageCheckMissingItemsOnOrdersEvent(t)

	// then
	wg.Wait()

	assert.Equal(t, 3, kitchenStubService.CalledCnt())
	assert.True(t, kitchenStubService.HaveBeenCalledWith(RequestMatchingFnc("hamburger", 1)))
	assert.True(t, kitchenStubService.HaveBeenCalledWith(RequestMatchingFnc("hamburger", 1)))
	assert.True(t, kitchenStubService.HaveBeenCalledWith(RequestMatchingFnc("cheeseburger", 2)))
}

func sendMessageCheckMissingItemsOnOrdersEvent(t *testing.T) {
	t.Log("Configure TestWriter for topic", orderJobsKafkaConfig.Topic)
	writer := event.NewTopicWriter(orderJobsKafkaConfig)

	msgKey := []byte(strconv.FormatInt(time.Now().UnixNano(), 10))

	headers := make([]kafka.Header, 0)
	headers = append(headers, kafka.Header{Key: "event", Value: []byte(CheckMissingItemsOnOrdersEvent)})

	msg := kafka.Message{
		Key:     msgKey,
		Headers: headers,
	}

	// when
	err := writer.SendMessage(context.Background(), msg)
	if err != nil {
		assert.Fail(t, "failed to send message to topic", orderJobsKafkaConfig.Topic)
	}
}
