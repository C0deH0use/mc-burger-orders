package main

import (
	"github.com/joho/godotenv"
	"github.com/segmentio/kafka-go"
	"mc-burger-orders/event"
	"mc-burger-orders/kitchen"
	"mc-burger-orders/log"
	"mc-burger-orders/middleware"
	"mc-burger-orders/order/management"
	"mc-burger-orders/schedule"
	"mc-burger-orders/shelf"
	sh "mc-burger-orders/shelf/handler"
)
import "github.com/gin-gonic/gin"
import "mc-burger-orders/order"

func main() {
	loadEnv()
	mongoDb := middleware.GetMongoClient()
	ordersShelf := shelf.NewEmptyShelf()
	eventBus := event.NewInternalEventBus()

	shelfTopicConfigs := shelf.TopicConfigsFromEnv()
	orderStatusTopicConfigs := order.StatusUpdatedTopicConfigsFromEnv()
	orderManagementJobsTopicConfigs := management.OrderJobsTopicConfigsFromEnv()
	orderStatusEndpointsTopicConfigs := order.StatusUpdatedEndpointTopicConfigsFromEnv()
	kitchenTopicConfigs := kitchen.TopicConfigsFromEnv()

	ordersShelf.ConfigureWriter(event.NewTopicWriter(shelfTopicConfigs))
	shelfHandlerTopicConfig := sh.TopicConfigsFromEnv()
	shelfHandler := sh.NewShelfHandler(kitchenTopicConfigs, ordersShelf)

	orderJobsReader := event.NewTopicReader(orderManagementJobsTopicConfigs, eventBus)
	orderStatusReader := event.NewTopicReader(orderStatusTopicConfigs, eventBus)

	stackTopicReader := event.NewTopicReader(shelfTopicConfigs, eventBus)
	shelfSchedulerReader := event.NewTopicReader(shelfHandlerTopicConfig, eventBus)

	orderCommandsHandler := order.NewHandler(mongoDb, kitchenTopicConfigs, orderStatusTopicConfigs, ordersShelf)
	orderManagementCommandsHandler := management.NewHandler(mongoDb, kitchenTopicConfigs)

	kitchenTopicReader := event.NewTopicReader(kitchenTopicConfigs, eventBus)
	kitchenEventsHandler := kitchen.NewHandler(ordersShelf)

	r := gin.Default()
	r.ForwardedByClientIP = true

	err := r.SetTrustedProxies([]string{"127.0.0.1"})
	if err != nil {
		log.Error.Panicf("error when setting trusted proxies. Reason: %s", err)
	}

	eventBus.AddHandler(shelfHandler)
	eventBus.AddHandler(orderCommandsHandler)
	eventBus.AddHandler(kitchenEventsHandler)
	eventBus.AddHandler(orderManagementCommandsHandler)

	orderEndpoints := order.NewOrderEndpoints(mongoDb, kitchenTopicConfigs, orderStatusTopicConfigs, ordersShelf)
	statusUpdatesEndpoints := order.NewOrderStatusEventsEndpoints(mongoDb, orderStatusEndpointsTopicConfigs)

	orderEndpoints.Setup(r)
	statusUpdatesEndpoints.Setup(r)

	go stackTopicReader.SubscribeToTopic(make(chan kafka.Message))
	go kitchenTopicReader.SubscribeToTopic(make(chan kafka.Message))
	go orderStatusReader.SubscribeToTopic(make(chan kafka.Message))
	go orderJobsReader.SubscribeToTopic(make(chan kafka.Message))
	go shelfSchedulerReader.SubscribeToTopic(make(chan kafka.Message))
	go schedule.ShelfJobs()
	go management.OrderManagementJobs()

	err = r.Run()

	if err != nil {
		log.Error.Panicf("error when starting REST service. Reason: %s", err)
	}
}

func loadEnv() {
	err := godotenv.Load()

	if err != nil {
		log.Error.Fatalf("Error loading .env file")
	}
}
