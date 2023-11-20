package order

import (
	"fmt"
	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/mongo"
	"mc-burger-orders/command"
	"mc-burger-orders/event"
	"mc-burger-orders/log"
	c "mc-burger-orders/order/command"
	m "mc-burger-orders/order/model"
	"mc-burger-orders/order/service"
	"mc-burger-orders/order/utils"
	"mc-burger-orders/stack"
	utils2 "mc-burger-orders/utils"
)

type CommandsHandler struct {
	command.DefaultCommandHandler
	stack          *stack.Stack
	queryService   m.OrderQueryService
	repository     m.OrderRepository
	kitchenService service.KitchenRequestService
}

func NewHandler(database *mongo.Database, kitchenTopicConfigs *event.TopicConfigs, s *stack.Stack) *CommandsHandler {
	repository := m.NewRepository(database)
	orderNumberRepository := m.NewOrderNumberRepository(database)
	queryService := m.OrderQueryService{Repository: repository, OrderNumberRepository: orderNumberRepository}
	kitchenService := service.NewKitchenServiceFrom(kitchenTopicConfigs)

	return &CommandsHandler{
		stack:          s,
		queryService:   queryService,
		repository:     repository,
		kitchenService: kitchenService,
	}
}

func (o *CommandsHandler) GetHandledEvents() []string {
	return []string{stack.ItemAddedToStackEvent, CollectedEvent}
}

func (o *CommandsHandler) GetCommands(message kafka.Message) ([]command.Command, error) {
	orderNumber, err := utils.GetOrderNumber(message)
	if err != nil {
		log.Error.Println(err.Error())
		return nil, err
	}

	eventType, err := utils2.GetEventType(message)
	if err != nil {
		log.Error.Println(err.Error())
		return nil, err
	}

	commands := make([]command.Command, 0)
	switch eventType {
	case stack.ItemAddedToStackEvent:
		{
			commands = append(commands, &c.PackItemCommand{
				Stack:          o.stack,
				Repository:     o.repository,
				KitchenService: o.kitchenService,
				Message:        message,
			})
		}
	case CollectedEvent:
		{
			commands = append(commands, &c.OrderCollectedCommand{
				Repository:  o.repository,
				OrderNumber: orderNumber,
			})
		}
	default:
		{
			err := fmt.Errorf("handling unknown event message: %s", eventType)
			log.Error.Println(err)
			return nil, err
		}
	}

	return commands, nil

}
