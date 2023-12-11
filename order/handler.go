package order

import (
	"fmt"
	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/mongo"
	"mc-burger-orders/command"
	"mc-burger-orders/event"
	"mc-burger-orders/log"
	"mc-burger-orders/shelf"
	utils2 "mc-burger-orders/utils"
)

type OrdersHandler struct {
	defaultHandler command.DefaultCommandHandler
	shelf          *shelf.Shelf
	queryService   OrderQueryService
	repository     OrderRepository
	statusEmitter  StatusEmitter
	kitchenService KitchenRequestService
}

func NewHandler(database *mongo.Database, kitchenTopicConfigs *event.TopicConfigs, statusEmitterTopicConfigs *event.TopicConfigs, s *shelf.Shelf) *OrdersHandler {
	repository := NewRepository(database)
	orderNumberRepository := NewOrderNumberRepository(database)
	queryService := OrderQueryService{Repository: repository, OrderNumberRepository: orderNumberRepository}
	kitchenService := NewKitchenServiceFrom(kitchenTopicConfigs)
	statusEmitter := NewStatusEmitterFrom(statusEmitterTopicConfigs)

	return &OrdersHandler{
		shelf:          s,
		queryService:   queryService,
		repository:     repository,
		kitchenService: kitchenService,
		statusEmitter:  statusEmitter,
		defaultHandler: command.DefaultCommandHandler{},
	}
}

func (o *OrdersHandler) Handle(message kafka.Message, commandResults chan command.TypedResult) {
	commands, err := o.GetCommands(message)
	if err != nil {
		commandResults <- command.NewErrorResult("OrderHandler", err)
		return
	}

	o.defaultHandler.HandleCommands(message, commandResults, commands...)
}

func (o *OrdersHandler) GetHandledEvents() []string {
	return []string{shelf.ItemAddedOnShelfEvent, StatusUpdatedEvent, CollectedEvent}
}

func (o *OrdersHandler) AddCommands(event string, commands ...command.Command) {
	o.defaultHandler.AddCommands(event, commands...)
}

func (o *OrdersHandler) GetCommands(message kafka.Message) ([]command.Command, error) {
	eventType, err := utils2.GetEventType(message)
	if err != nil {
		log.Error.Println(err.Error())
		return nil, err
	}

	commands := make([]command.Command, 0)
	switch eventType {
	case shelf.ItemAddedOnShelfEvent:
		{
			commands = append(commands, &PackItemCommand{
				Shelf:          o.shelf,
				Repository:     o.repository,
				KitchenService: o.kitchenService,
				StatusEmitter:  o.statusEmitter,
			})
		}
	case StatusUpdatedEvent:
		{
			orderNumber, err := utils2.GetOrderNumber(message)
			if err != nil {
				log.Error.Println(err.Error())
				return nil, err
			}
			commands = append(commands, &OrderUpdatedCommand{
				Repository:  o.repository,
				OrderNumber: orderNumber,
			})
		}
	case CollectedEvent:
		{
			orderNumber, err := utils2.GetOrderNumber(message)
			if err != nil {
				log.Error.Println(err.Error())
				return nil, err
			}
			commands = append(commands, &OrderCollectedCommand{
				Repository:  o.repository,
				OrderNumber: orderNumber,
			})
		}
	default:
		{
			err := fmt.Errorf("handling unknown event message: %s", eventType)
			log.Error.Println(err)
		}
	}

	return commands, nil
}
