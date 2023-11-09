package order

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"mc-burger-orders/command"
	i "mc-burger-orders/item"
	"mc-burger-orders/log"
	"mc-burger-orders/middleware"
	command2 "mc-burger-orders/order/command"
	m "mc-burger-orders/order/model"
	"mc-burger-orders/order/service"
	"mc-burger-orders/stack"
	"mc-burger-orders/utils"
	"net/http"
)

type Endpoints struct {
	stack           *stack.Stack
	queryService    m.OrderQueryService
	orderRepository m.OrderRepository
	kitchenService  service.KitchenRequestService
	commandHandler  command.ExecutionHandler
}

func NewOrderEndpoints(database *mongo.Database, kitchenConfigs service.KitchenServiceConfigs, executorHandler command.ExecutionHandler) middleware.EndpointsSetup {
	s := stack.NewStack(stack.CleanStack())
	repository := m.NewRepository(database)
	orderNumberRepository := m.NewOrderNumberRepository(database)
	queryService := m.OrderQueryService{Repository: repository, OrderNumberRepository: orderNumberRepository}
	kitchenService := service.NewKitchenServiceFrom(kitchenConfigs)
	handler := executorHandler

	return &Endpoints{
		stack: s, queryService: queryService, orderRepository: repository, kitchenService: kitchenService, commandHandler: handler,
	}
}

func (e *Endpoints) CreateNewOrderCommand(orderNumber int64, order m.NewOrder) command.Command {
	return &command2.NewRequestCommand{
		Stack:          e.stack,
		Repository:     e.orderRepository,
		KitchenService: e.kitchenService,
		OrderNumber:    orderNumber,
		NewOrder:       order,
	}
}

func (e *Endpoints) Setup(r *gin.Engine) {
	r.GET("/order", e.queryService.FetchOrders)
	r.PUT("/order", e.newOrderHandler)
}

func (e *Endpoints) newOrderHandler(c *gin.Context) {
	newOrder := m.NewOrder{}
	err := c.ShouldBindJSON(&newOrder)
	if err != nil {
		errorMessage := fmt.Sprintf("Schema Error. %s", err)
		log.Info.Println("New Order request Error: ", errorMessage)
		c.JSON(http.StatusBadRequest, utils.ErrorPayload(errorMessage))
		return
	}
	err = validate(newOrder)
	if err != nil {
		log.Error.Println(err)
		c.JSON(http.StatusBadRequest, utils.ErrorPayload(err.Error()))
		return
	}

	orderNumber := e.queryService.GetNextOrderNumber(c)
	result, err := e.commandHandler.Execute(e.CreateNewOrderCommand(orderNumber, newOrder))

	if err != nil {
		log.Error.Println(err)
		c.JSON(http.StatusBadRequest, utils.ErrorPayload(err.Error()))
		return
	}
	if !result {
		c.JSON(http.StatusBadRequest, utils.ErrorPayload("Failed to Create new order"))
		return
	}

	c.JSON(http.StatusCreated, map[string]int64{"orderNumber": orderNumber})
}

func validate(c m.NewOrder) error {
	var errs []error
	for _, item := range c.Items {
		err := i.KnownItem(item.Name)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}
