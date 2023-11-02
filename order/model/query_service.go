package model

import (
	"github.com/gin-gonic/gin"
	"log"
	"mc-burger-orders/utils"
	"net/http"
)

type OrderQueryService struct {
	Repository *OrderRepository
}

func (s *OrderQueryService) FetchOrders(c *gin.Context) {
	orders, err := s.Repository.FetchMany(c)

	if err != nil {
		log.Println("Failure when reading data from db.", err.Error())

		c.JSON(http.StatusInternalServerError, utils.ErrorPayload(err.Error()))
		return
	}

	c.JSON(http.StatusOK, orders)
}
