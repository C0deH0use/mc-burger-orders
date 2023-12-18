package order

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/mongo"
	"mc-burger-orders/command"
	"mc-burger-orders/event"
	i "mc-burger-orders/kitchen/item"
	"mc-burger-orders/log"
	"mc-burger-orders/middleware"
	"mc-burger-orders/shelf"
	"mc-burger-orders/testing/utils"
	"net/http"
	"strconv"
)

type Endpoints struct {
	stack           *shelf.Shelf
	queryService    OrderQueryService
	orderRepository OrderRepository
	kitchenService  KitchenRequestService
	statusEmitter   StatusEmitter
	dispatcher      command.Dispatcher
}

func NewOrderEndpoints(database *mongo.Database, kitchenTopicConfigs *event.TopicConfigs, statusEmitterTopicConfigs *event.TopicConfigs, streamService OrderStreamService, s *shelf.Shelf) middleware.EndpointsSetup {
	repository := NewRepository(database, streamService)
	orderNumberRepository := NewOrderNumberRepository(database)
	queryService := OrderQueryService{Repository: repository, orderNumberRepository: orderNumberRepository}
	kitchenService := NewKitchenServiceFrom(kitchenTopicConfigs)
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

func (e *Endpoints) CreateNewOrderCommand(orderNumber int64, order NewOrder) command.Command {
	return &NewRequestCommand{
		Shelf:          e.stack,
		Repository:     e.orderRepository,
		KitchenService: e.kitchenService,
		StatusEmitter:  e.statusEmitter,
		OrderNumber:    orderNumber,
		NewOrder:       order,
	}
}

func (e *Endpoints) Setup(r *gin.Engine) {
	r.GET("/order", e.queryService.FetchOrders)
	r.POST("/order", e.newOrderHandler)
	r.POST("/order/:orderNumber/collect", e.collectOrderHandler)
}

func (e *Endpoints) newOrderHandler(c *gin.Context) {
	newOrder := NewOrder{}
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

	commandResults := make(chan command.TypedResult)
	orderNumber := e.queryService.GetNextOrderNumber(c)
	cmd := e.CreateNewOrderCommand(orderNumber, newOrder)
	go e.dispatcher.Execute(cmd, kafka.Message{}, commandResults)

	commandResult := <-commandResults

	if commandResult.Error != nil {
		log.Error.Println(commandResult.Error.ErrorMessage)
		c.JSON(commandResult.Error.HttpResponse, utils.ErrorPayload(commandResult.Error.ErrorMessage))
		return
	}

	c.JSON(http.StatusCreated, map[string]int64{"orderNumber": orderNumber})
}

func (e *Endpoints) collectOrderHandler(c *gin.Context) {
	param := c.Param("orderNumber")
	orderNumber, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		log.Error.Println("Unable to parse url parameter to OrderNumber", err.Error())

		errResponse := fmt.Sprintf("unable to parse url parameter to OrderNumber. Reason - %v", err.Error())
		c.JSON(http.StatusBadRequest, utils.ErrorPayload(errResponse))
		return
	}

	commandResults := make(chan command.TypedResult)
	cmd := &OrderCollectedCommand{OrderNumber: orderNumber, Repository: e.orderRepository, StatusEmitter: e.statusEmitter}
	go e.dispatcher.Execute(cmd, kafka.Message{}, commandResults)

	commandResult := <-commandResults

	if commandResult.Error != nil {
		log.Error.Println(commandResult.Error.ErrorMessage)
		c.JSON(commandResult.Error.HttpResponse, utils.ErrorPayload(commandResult.Error.ErrorMessage))
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func validate(c NewOrder) error {
	var errs []error
	for _, item := range c.Items {
		if err := i.IsKnownItem(item.Name); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}
