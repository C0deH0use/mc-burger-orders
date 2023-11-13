package event

import (
	"github.com/segmentio/kafka-go"
	"mc-burger-orders/command"
)

type EventBus interface {
	//FIXME: change kafka.Message to something internal
	PublishEvent(message kafka.Message) error

	AddHandler(command.Handler, ...string)
}
