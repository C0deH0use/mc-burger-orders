package main

import (
	"github.com/joho/godotenv"
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
	kitchenServiceConfigs := service.KitchenServiceConfigsFromEnv()
	eventBus := event.NewInternalEventBus()

	orderEventHandler := order.NewOrderEventHandler(mongoDb, kitchenServiceConfigs, kitchenStack)
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

	orderEndpoints := order.NewOrderEndpoints(mongoDb, kitchenServiceConfigs, kitchenStack)

	orderEndpoints.Setup(r)

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
