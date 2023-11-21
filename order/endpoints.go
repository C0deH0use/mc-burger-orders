package order

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"mc-burger-orders/command"
	"mc-burger-orders/event"
	i "mc-burger-orders/kitchen/item"
	"mc-burger-orders/log"
	"mc-burger-orders/middleware"
	m "mc-burger-orders/order/model"
	"mc-burger-orders/order/service"
	"mc-burger-orders/stack"
	"mc-burger-orders/testing/utils"
	"net/http"
)

type Endpoints struct {
	stack           *stack.Stack
	queryService    m.OrderQueryService
	orderRepository m.OrderRepository
	kitchenService  service.KitchenRequestService
	statusEmitter   StatusEmitter
	dispatcher      command.Dispatcher
}

func NewOrderEndpoints(database *mongo.Database, kitchenTopicConfigs *event.TopicConfigs, statusEmitterTopicConfigs *event.TopicConfigs, s *stack.Stack) middleware.EndpointsSetup {
	repository := m.NewRepository(database)
	orderNumberRepository := m.NewOrderNumberRepository(database)
	queryService := m.OrderQueryService{Repository: repository, OrderNumberRepository: orderNumberRepository}
	kitchenService := service.NewKitchenServiceFrom(kitchenTopicConfigs)
	statusEmitter := NewStatusEmitterFrom(statusEmitterTopicConfigs)

	return &Endpoints{
		stack:           s,
		queryService:    queryService,
		orderRepository: repository,
		kitchenService:  kitchenService,
		statusEmitter:   statusEmitter,
		dispatcher:      &command.DefaultDispatcher{},
	}
}

func (e *Endpoints) CreateNewOrderCommand(orderNumber int64, order m.NewOrder) command.Command {
	return &NewRequestCommand{
		Stack:          e.stack,
		Repository:     e.orderRepository,
		KitchenService: e.kitchenService,
		StatusEmitter:  e.statusEmitter,
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
	cmd := e.CreateNewOrderCommand(orderNumber, newOrder)
	result, err := e.dispatcher.Execute(cmd)

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
		if err := i.IsKnownItem(item.Name); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}
