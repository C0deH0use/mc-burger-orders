package main

import (
	"github.com/joho/godotenv"
	"github.com/segmentio/kafka-go"
	"mc-burger-orders/event"
	"mc-burger-orders/kitchen"
	"mc-burger-orders/log"
	"mc-burger-orders/middleware"
	"mc-burger-orders/stack"
)
import "github.com/gin-gonic/gin"
import "mc-burger-orders/order"

func main() {
	loadEnv()
	mongoDb := middleware.GetMongoClient()
	kitchenStack := stack.NewEmptyStack()
	eventBus := event.NewInternalEventBus()

	stackTopicConfigs := stack.TopicConfigsFromEnv()
	orderStatusTopicConfigs := order.StatusUpdatedTopicConfigsFromEnv()
	kitchenTopicConfigs := kitchen.TopicConfigsFromEnv()

	kitchenStack.ConfigureWriter(event.NewTopicWriter(stackTopicConfigs))
	stackTopicReader := event.NewTopicReader(stackTopicConfigs, eventBus)
	orderStatusReader := event.NewTopicReader(orderStatusTopicConfigs, eventBus)

	orderCommandsHandler := order.NewHandler(mongoDb, kitchenTopicConfigs, orderStatusTopicConfigs, kitchenStack)

	kitchenTopicReader := event.NewTopicReader(kitchenTopicConfigs, eventBus)
	kitchenEventsHandler := kitchen.NewHandler(kitchenStack)

	r := gin.Default()
	r.ForwardedByClientIP = true

	err := r.SetTrustedProxies([]string{"127.0.0.1"})
	if err != nil {
		log.Error.Panicf("error when setting trusted proxies. Reason: %s", err)
	}

	eventBus.AddHandler(orderCommandsHandler)
	eventBus.AddHandler(kitchenEventsHandler)

	orderEndpoints := order.NewOrderEndpoints(mongoDb, kitchenTopicConfigs, orderStatusTopicConfigs, kitchenStack)

	orderEndpoints.Setup(r)

	go stackTopicReader.SubscribeToTopic(make(chan kafka.Message))
	go kitchenTopicReader.SubscribeToTopic(make(chan kafka.Message))
	go orderStatusReader.SubscribeToTopic(make(chan kafka.Message))

	err = r.Run()

	if err != nil {
		log.Error.Panicf("error when starting REST service. Reason: %s", err)
	}
}

func loadEnv() {
	// load .env file
	err := godotenv.Load()

	if err != nil {
		log.Error.Fatalf("Error loading .env file")
	}
}
