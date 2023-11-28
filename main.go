package main

import (
	"github.com/joho/godotenv"
	"github.com/segmentio/kafka-go"
	"mc-burger-orders/event"
	"mc-burger-orders/kitchen"
	"mc-burger-orders/log"
	"mc-burger-orders/middleware"
	"mc-burger-orders/shelf"
	sh "mc-burger-orders/shelf/handler"
	"time"
)
import "github.com/gin-gonic/gin"
import "mc-burger-orders/order"

func main() {
	loadEnv()
	mongoDb := middleware.GetMongoClient()
	ordersShelf := shelf.NewEmptyShelf()
	eventBus := event.NewInternalEventBus()

	stackTopicConfigs := shelf.TopicConfigsFromEnv()
	orderStatusTopicConfigs := order.StatusUpdatedTopicConfigsFromEnv()
	kitchenTopicConfigs := kitchen.TopicConfigsFromEnv()

	ordersShelf.ConfigureWriter(event.NewTopicWriter(stackTopicConfigs))
	shelfHandler := sh.NewShelfHandler(kitchenTopicConfigs, ordersShelf)
	eventBus.AddHandler(shelfHandler)

	stackTopicReader := event.NewTopicReader(stackTopicConfigs, eventBus)
	orderStatusReader := event.NewTopicReader(orderStatusTopicConfigs, eventBus)

	orderCommandsHandler := order.NewHandler(mongoDb, kitchenTopicConfigs, orderStatusTopicConfigs, ordersShelf)

	kitchenTopicReader := event.NewTopicReader(kitchenTopicConfigs, eventBus)
	kitchenEventsHandler := kitchen.NewHandler(ordersShelf)

	r := gin.Default()
	r.ForwardedByClientIP = true

	err := r.SetTrustedProxies([]string{"127.0.0.1"})
	if err != nil {
		log.Error.Panicf("error when setting trusted proxies. Reason: %s", err)
	}

	eventBus.AddHandler(orderCommandsHandler)
	eventBus.AddHandler(kitchenEventsHandler)

	orderEndpoints := order.NewOrderEndpoints(mongoDb, kitchenTopicConfigs, orderStatusTopicConfigs, ordersShelf)

	orderEndpoints.Setup(r)

	ScheduleJobs(eventBus, shelfHandler)
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

func ScheduleJobs(bus event.EventBus, shelfHandler *sh.Handler) {
	go func() {
		for {
			go func() {
				err := bus.PublishEvent(shelfHandler.CheckFavoritesOnShelfMessage())
				if err != nil {
					log.Error.Println("failed to publish message on event bus", err)
				}
			}()

			time.Sleep(3 * time.Minute)
		}
	}()
}
