package main

import (
	"fmt"
	"net/http"
)
import "github.com/gin-gonic/gin"
import "mc-burger-orders/order"

func main() {
	r := gin.New()
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello World!",
		})
	})
	orderEndpoints := order.NewOrderEndpoints()

	orderEndpoints.Setup(r)

	err := r.Run()

	if err != nil {
		fmt.Println("Error when starting REST Service", err)
	}
}
