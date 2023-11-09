package main

import (
	"github.com/joho/godotenv"
	"log"
	"mc-burger-orders/command"
	"mc-burger-orders/middleware"
	"net/http"
)
import "github.com/gin-gonic/gin"
import "mc-burger-orders/order"

func main() {
	loadEnv()
	mongoDb := middleware.GetMongoClient()
	executorHandler := &command.DefaultHandler{}

	r := gin.Default()
	r.ForwardedByClientIP = true
	err := r.SetTrustedProxies([]string{"127.0.0.1"})
	if err != nil {
		log.Panicf("error when setting trusted proxies. Reason: %s", err)
	}

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello World!",
		})
	})

	//executorHandler.Register("CreateOrder", []command.Command{orderCommands.NewRequestCommand})

	orderEndpoints := order.NewOrderEndpoints(mongoDb, executorHandler)

	orderEndpoints.Setup(r)

	err = r.Run()

	if err != nil {
		log.Panicf("error when starting REST service. Reason: %s", err)
	}
}

func loadEnv() {
	// load .env file
	err := godotenv.Load()

	if err != nil {
		log.Fatalf("Error loading .env file")
	}
}
