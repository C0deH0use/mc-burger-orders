package main

import (
	"github.com/joho/godotenv"
	"mc-burger-orders/command"
	"mc-burger-orders/log"
	"mc-burger-orders/middleware"
	"mc-burger-orders/order/service"
	"net/http"
)
import "github.com/gin-gonic/gin"
import "mc-burger-orders/order"

func main() {
	mongoDb := middleware.GetMongoClient()
	executorHandler := &command.DefaultHandler{}
	kitchenServiceConfigs := service.KitchenServiceConfigsFromEnv()

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

	//executorHandler.Register("CreateOrder", []command.Command{orderCommands.NewRequestCommand})

	orderEndpoints := order.NewOrderEndpoints(mongoDb, kitchenServiceConfigs, executorHandler)

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
