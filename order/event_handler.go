package order

import (
	"fmt"
	"github.com/segmentio/kafka-go"
	"github.com/spf13/cast"
	"go.mongodb.org/mongo-driver/mongo"
	"mc-burger-orders/command"
	"mc-burger-orders/event"
	"mc-burger-orders/log"
	c "mc-burger-orders/order/command"
	m "mc-burger-orders/order/model"
	"mc-burger-orders/order/service"
	"mc-burger-orders/stack"
)

var (
	CollectedEvent = "order-collected"
)

type EventHandler struct {
	event.DefaultEventHandler
	stack          *stack.Stack
	queryService   m.OrderQueryService
	repository     m.OrderRepository
	kitchenService service.KitchenRequestService
}

func NewOrderEventHandler(database *mongo.Database, kitchenTopicConfigs *event.TopicConfigs, s *stack.Stack) *EventHandler {
	repository := m.NewRepository(database)
	orderNumberRepository := m.NewOrderNumberRepository(database)
	queryService := m.OrderQueryService{Repository: repository, OrderNumberRepository: orderNumberRepository}
	kitchenService := service.NewKitchenServiceFrom(kitchenTopicConfigs)

	return &EventHandler{
		stack:          s,
		queryService:   queryService,
		repository:     repository,
		kitchenService: kitchenService,
	}
}

func (o *EventHandler) GetCommand(message kafka.Message) (command.Command, error) {
	var orderNumber int64
	for _, header := range message.Headers {
		if header.Key == "order" {
			orderNumberStr := string(header.Value)
			orderNumber = cast.ToInt64(orderNumberStr)
		}
	}

	eventType := string(message.Key)
	switch eventType {
	case stack.ItemAddedToStackEvent:
		{
			return &c.PackItemCommand{
				Stack:       o.stack,
				Repository:  o.repository,
				OrderNumber: orderNumber,
			}, nil
		}
	case CollectedEvent:
		{
			return &c.OrderCollectedCommand{
				Repository:  o.repository,
				OrderNumber: orderNumber,
			}, nil
		}
	default:
		{
			err := fmt.Errorf("handling unknown event message: %s", eventType)
			log.Error.Println(err)
			return nil, err
		}
	}
}
