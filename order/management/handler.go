package management

import (
	"fmt"
	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/mongo"
	"mc-burger-orders/command"
	"mc-burger-orders/event"
	"mc-burger-orders/log"
	"mc-burger-orders/order"
	utils2 "mc-burger-orders/utils"
)

type OrderManagementHandler struct {
	defaultHandler command.DefaultCommandHandler
	queryService   OrderQueryService
	kitchenService order.KitchenRequestService
}

func NewHandler(database *mongo.Database, kitchenTopicConfigs *event.TopicConfigs) *OrderManagementHandler {
	queryService := NewOrderQueryService(NewOrderRepository(database))
	kitchenService := order.NewKitchenServiceFrom(kitchenTopicConfigs)

	return &OrderManagementHandler{
		queryService:   queryService,
		kitchenService: kitchenService,
		defaultHandler: command.DefaultCommandHandler{},
	}
}

func (o *OrderManagementHandler) Handle(message kafka.Message, commandResults chan command.TypedResult) {
	commands, err := o.GetCommands(message)
	if err != nil {
		commandResults <- command.NewErrorResult("OrderManagementHandler", err)
		return
	}

	o.defaultHandler.HandleCommands(message, commandResults, commands...)
}

func (o *OrderManagementHandler) GetHandledEvents() []string {
	return []string{CheckMissingItemsOnOrdersEvent}
}

func (o *OrderManagementHandler) AddCommands(event string, commands ...command.Command) {
	o.defaultHandler.AddCommands(event, commands...)
}

func (o *OrderManagementHandler) GetCommands(message kafka.Message) ([]command.Command, error) {
	eventType, err := utils2.GetEventType(message)
	if err != nil {
		log.Error.Println(err.Error())
		return nil, err
	}

	commands := make([]command.Command, 0)
	switch eventType {
	case CheckMissingItemsOnOrdersEvent:
		{
			commands = append(commands, &CheckMissingItemsOnOrdersCommand{
				queryService:   o.queryService,
				kitchenService: o.kitchenService,
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
