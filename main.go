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

var (
	mongoDb         = middleware.GetMongoClient()
	executorHandler = &command.DefaultHandler{}
)

func main() {
	loadEnv()
	r := gin.New()
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello World!",
		})
	})

	//executorHandler.Register("CreateOrder", []command.Command{orderCommands.NewRequestCommand})

	orderEndpoints := order.NewOrderEndpoints(mongoDb, executorHandler)

	orderEndpoints.Setup(r)

	err := r.Run()

	if err != nil {
		log.Println("Error when starting REST Service", err)
	}
}

func loadEnv() {
	// load .env file
	err := godotenv.Load()

	if err != nil {
		log.Fatalf("Error loading .env file")
	}
}
