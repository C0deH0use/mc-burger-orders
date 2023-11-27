package order

import (
	"fmt"
	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/mongo"
	"mc-burger-orders/command"
	"mc-burger-orders/event"
	"mc-burger-orders/log"
	"mc-burger-orders/stack"
	utils2 "mc-burger-orders/utils"
)

type OrdersHandler struct {
	defaultHandler command.DefaultCommandHandler
	stack          *stack.Stack
	queryService   OrderQueryService
	repository     OrderRepository
	statusEmitter  StatusEmitter
	kitchenService KitchenRequestService
}

func NewHandler(database *mongo.Database, kitchenTopicConfigs *event.TopicConfigs, statusEmitterTopicConfigs *event.TopicConfigs, s *stack.Stack) *OrdersHandler {
	repository := NewRepository(database)
	orderNumberRepository := NewOrderNumberRepository(database)
	queryService := OrderQueryService{Repository: repository, OrderNumberRepository: orderNumberRepository}
	kitchenService := NewKitchenServiceFrom(kitchenTopicConfigs)
	statusEmitter := NewStatusEmitterFrom(statusEmitterTopicConfigs)

	return &OrdersHandler{
		stack:          s,
		queryService:   queryService,
		repository:     repository,
		kitchenService: kitchenService,
		statusEmitter:  statusEmitter,
		defaultHandler: command.DefaultCommandHandler{},
	}
}

func (o *OrdersHandler) Handle(message kafka.Message) (bool, error) {
	commands, err := o.GetCommands(message)
	if err != nil {
		log.Error.Println(err.Error())
		return false, err
	}

	return o.defaultHandler.HandleCommands(message, commands...)
}

func (o *OrdersHandler) GetHandledEvents() []string {
	return []string{stack.ItemAddedToStackEvent, StatusUpdatedEvent, CollectedEvent}
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
	case stack.ItemAddedToStackEvent:
		{
			commands = append(commands, &PackItemCommand{
				Stack:          o.stack,
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
