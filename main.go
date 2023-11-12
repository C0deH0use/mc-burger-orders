package main

import (
	"context"
	"github.com/joho/godotenv"
	"github.com/segmentio/kafka-go"
	"mc-burger-orders/event"
	"mc-burger-orders/log"
	"mc-burger-orders/middleware"
	"mc-burger-orders/order/service"
	"mc-burger-orders/stack"
	"net/http"
)
import "github.com/gin-gonic/gin"
import "mc-burger-orders/order"

func main() {
	loadEnv()
	mongoDb := middleware.GetMongoClient()
	kitchenStack := stack.NewStack(stack.CleanStack())
	stackTopicConfigs := stack.TopicConfigsFromEnv()
	kitchenTopicConfigs := service.KitchenTopicConfigsFromEnv()
	stackTopicReader := event.NewTopicReader(stackTopicConfigs)
	eventBus := event.NewInternalEventBus()

	orderEventHandler := order.NewOrderEventHandler(mongoDb, kitchenTopicConfigs, kitchenStack)
	r := gin.Default()
	r.ForwardedByClientIP = true
	err := r.SetTrustedProxies([]string{"127.0.0.1"})
	if err != nil {
		log.Error.Panicf("error when setting trusted proxies. Reason: %s", err)
	}

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello World!",
		})
	})

	eventBus.AddHandler(orderEventHandler, stack.ItemAddedToStackEvent, order.CollectedEvent)

	orderEndpoints := order.NewOrderEndpoints(mongoDb, kitchenTopicConfigs, kitchenStack)

	orderEndpoints.Setup(r)

	err = r.Run()

	if err != nil {
		log.Error.Panicf("error when starting REST service. Reason: %s", err)
	}

	stackMessages := make(chan kafka.Message)
	go stackTopicReader.SubscribeToTopic(context.Background(), stackMessages)

	go func() {
		for {
			select {
			case newMessage := <-stackMessages:
				{
					messageKey := string(newMessage.Key)
					message := string(newMessage.Value)
					log.Info.Println("New message read from topic (", newMessage.Topic, ") arrived, key: ", messageKey, "message value:", message)

					err := eventBus.PublishEvent(newMessage)
					if err != nil {
						log.Error.Panicf("failed to publish message on event bus", err)
					}
				}
			}
		}
	}()
}

func loadEnv() {
	// load .env file
	err := godotenv.Load()

	if err != nil {
		log.Error.Fatalf("Error loading .env file")
	}
}
