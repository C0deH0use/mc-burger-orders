package event

import (
	"github.com/segmentio/kafka-go"
)

type EventBus interface {
	//FIXME: change kafka.Message to something internal
	PublishEvent(message kafka.Message) error

	AddHandler(Handler, ...interface{})
}
