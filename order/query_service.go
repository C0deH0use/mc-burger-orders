package order

import (
	"github.com/gin-gonic/gin"
	"mc-burger-orders/log"
	"mc-burger-orders/testing/utils"
	"net/http"
)

type OrderQueryService struct {
	OrderNumberRepository OrderNumberRepository
	Repository            FetchManyRepository
}

func (s *OrderQueryService) FetchOrders(c *gin.Context) {
	orders, err := s.Repository.FetchMany(c)

	if err != nil {
		log.Error.Println("Failure when reading data from db.", err.Error())

		c.JSON(http.StatusInternalServerError, utils.ErrorPayload(err.Error()))
		return
	}

	c.JSON(http.StatusOK, orders)
}

func (s *OrderQueryService) GetNextOrderNumber(c *gin.Context) int64 {
	orderNumber, err := s.OrderNumberRepository.GetNext(c)

	if err != nil {
		log.Error.Println("Failure when Next Order Number from db.", err.Error())

		c.JSON(http.StatusInternalServerError, utils.ErrorPayload(err.Error()))
		return 0
	}

	return orderNumber
}
