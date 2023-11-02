package order

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	i "mc-burger-orders/item"
	"mc-burger-orders/middleware"
	command2 "mc-burger-orders/order/command"
	m "mc-burger-orders/order/model"
	"mc-burger-orders/order/service"
	"mc-burger-orders/stack"
	"mc-burger-orders/utils"
	"net/http"
)

type Endpoints struct {
	stack          *stack.Stack
	queryService   m.OrderQueryService
	repository     *m.OrderRepository
	kitchenService service.KitchenRequestService
	commandHandler command2.CommandHandler
}

func NewOrderEndpoints(database *mongo.Database) middleware.EndpointsSetup {
	s := stack.NewStack(stack.CleanStack())
	repository := m.NewRepository(database)
	queryService := m.OrderQueryService{Repository: repository}
	kitchenService := &service.KitchenService{}
	handler := &command2.DefaultHandler{}

	return &Endpoints{
		stack: s, queryService: queryService, repository: repository, kitchenService: kitchenService, commandHandler: handler,
	}
}

func (e *Endpoints) CreateNewOrderCommand(order m.NewOrder) command2.Command {
	return &command2.NewRequestCommand{
		Stack:          e.stack,
		Repository:     e.repository,
		KitchenService: e.kitchenService,
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
		log.Println("New Order request Error: ", errorMessage)
		c.JSON(http.StatusBadRequest, utils.ErrorPayload(errorMessage))
		return
	}
	err = validate(newOrder)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, utils.ErrorPayload(err.Error()))
		return
	}

	command := e.CreateNewOrderCommand(newOrder)
	order, err := e.commandHandler.Execute(command)

	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, utils.ErrorPayload(err.Error()))
		return
	}

	c.JSON(http.StatusCreated, order)
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
