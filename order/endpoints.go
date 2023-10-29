package order

import (
	"github.com/gin-gonic/gin"
	"mc-burger-orders/order/command"
	"mc-burger-orders/order/model"
	"mc-burger-orders/stack"
)

func Endpoints(r *gin.Engine) {
	s := stack.NewStack(stack.CleanStack())
	queryService := model.OrderQueryService{Repository: model.OrderRepository{}}
	newOrderCommand := command.NewOrderRequestCommand{Stack: s, Repository: model.OrderRepository{}}

	r.GET("/order", queryService.FetchOrders)
	r.PUT("/order", newOrderCommand.CreateNewOrder)
}
