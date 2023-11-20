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
	kitchenTopicConfigs := kitchen.TopicConfigsFromEnv()
	stackTopicReader := event.NewTopicReader(stackTopicConfigs, eventBus)

	orderCommandsHandler := order.NewHandler(mongoDb, kitchenTopicConfigs, kitchenStack)

	kitchenTopicReader := event.NewTopicReader(kitchenTopicConfigs, eventBus)
	kitchenEventsHandler := kitchen.NewHandler(kitchenTopicConfigs, stackTopicConfigs, kitchenStack)

	r := gin.Default()
	r.ForwardedByClientIP = true

	err := r.SetTrustedProxies([]string{"127.0.0.1"})
	if err != nil {
		log.Error.Panicf("error when setting trusted proxies. Reason: %s", err)
	}

	eventBus.AddHandler(orderCommandsHandler)
	eventBus.AddHandler(kitchenEventsHandler)

	orderEndpoints := order.NewOrderEndpoints(mongoDb, kitchenTopicConfigs, kitchenStack)

	orderEndpoints.Setup(r)

	err = r.Run()

	if err != nil {
		log.Error.Panicf("error when starting REST service. Reason: %s", err)
	}

	go stackTopicReader.SubscribeToTopic(make(chan kafka.Message))
	go kitchenTopicReader.SubscribeToTopic(make(chan kafka.Message))

}

func loadEnv() {
	// load .env file
	err := godotenv.Load()

	if err != nil {
		log.Error.Fatalf("Error loading .env file")
	}
}
