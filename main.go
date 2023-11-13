package main

import (
	"github.com/joho/godotenv"
	"github.com/segmentio/kafka-go"
	"mc-burger-orders/event"
	"mc-burger-orders/log"
	"mc-burger-orders/middleware"
	"mc-burger-orders/order/service"
	"mc-burger-orders/stack"
)
import "github.com/gin-gonic/gin"
import "mc-burger-orders/order"

func main() {
	loadEnv()
	mongoDb := middleware.GetMongoClient()
	kitchenStack := stack.NewStack(stack.CleanStack())
	stackTopicConfigs := stack.TopicConfigsFromEnv()
	kitchenTopicConfigs := service.KitchenTopicConfigsFromEnv()
	eventBus := event.NewInternalEventBus()
	stackTopicReader := event.NewTopicReader(stackTopicConfigs, eventBus)

	orderCommandsHandler := order.NewOrderCommandHandler(mongoDb, kitchenTopicConfigs, kitchenStack)
	stackMessages := make(chan kafka.Message)

	r := gin.Default()
	r.ForwardedByClientIP = true

	err := r.SetTrustedProxies([]string{"127.0.0.1"})
	if err != nil {
		log.Error.Panicf("error when setting trusted proxies. Reason: %s", err)
	}

	eventBus.AddHandler(orderCommandsHandler, stack.ItemAddedToStackEvent, order.CollectedEvent)

	orderEndpoints := order.NewOrderEndpoints(mongoDb, kitchenTopicConfigs, kitchenStack)

	orderEndpoints.Setup(r)

	err = r.Run()

	if err != nil {
		log.Error.Panicf("error when starting REST service. Reason: %s", err)
	}

	stackTopicReader.SubscribeToTopic(stackMessages)

}

func loadEnv() {
	// load .env file
	err := godotenv.Load()

	if err != nil {
		log.Error.Fatalf("Error loading .env file")
	}
}
