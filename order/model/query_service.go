package model

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type OrderQueryService struct {
	Repository OrderRepository
}

func (s *OrderQueryService) FetchOrders(c *gin.Context) {
	orders := s.Repository.FetchOrders()

	c.JSON(http.StatusOK, orders)
}
