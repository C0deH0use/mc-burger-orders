package event

import (
	"github.com/segmentio/kafka-go"
	"mc-burger-orders/command"
)

type EventBus interface {
	PublishEvent(message kafka.Message, commandsResult chan command.TypedResult)

	AddHandler(command.Handler)
}
