package stack

import (
	"fmt"
	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/mongo"
	"mc-burger-orders/command"
	"mc-burger-orders/log"
)

var (
	ItemAddedToStackEvent = "item-added-to-stack"
)

type EventHandler struct {
	command.DefaultCommandHandler
	stack *Stack
}

func NewStackEventHandler(database *mongo.Database, s *Stack) *EventHandler {
	return &EventHandler{stack: s}
}

func (o *EventHandler) GetCommand(message kafka.Message) (command.Command, error) {
	eventType := message.Topic
	switch eventType {
	case ItemAddedToStackEvent:
		return nil, nil
	default:
		{
			err := fmt.Errorf("handling unknown event message: %s", eventType)
			log.Error.Println(err)
			return nil, err
		}
	}
}
