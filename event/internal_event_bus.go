package event

import (
	"github.com/segmentio/kafka-go"
	"mc-burger-orders/command"
	"mc-burger-orders/log"
	"mc-burger-orders/utils"
	"reflect"
)

type InternalEventBus struct {
	command.DefaultCommandHandler
	eventHandlers map[string]map[command.Handler]struct{}
}

func NewInternalEventBus() *InternalEventBus {
	return &InternalEventBus{
		eventHandlers: make(map[string]map[command.Handler]struct{}),
	}
}

func (b *InternalEventBus) PublishEvent(message kafka.Message) error {
	eventType, err := utils.GetEventType(message)
	if err != nil {
		log.Error.Println(err.Error())
		return err
	}

	if handlers, ok := b.eventHandlers[eventType]; ok {
		for handler := range handlers {
			result, err := handler.Handle(message)
			if err != nil {
				log.Error.Printf("Error when processing following event: %v by handler -> %v\n", eventType, reflect.TypeOf(handler))
				return err
			}
			log.Info.Printf("Event %v was handled with result %v\n", eventType, result)
		}
		return nil
	}
	log.Error.Printf("Event %v does not have any handled", eventType)
	return nil
}

func (b *InternalEventBus) AddHandler(handler command.Handler) {

	for _, event := range handler.GetHandledEvents() {

		if _, ok := b.eventHandlers[event]; !ok {
			b.eventHandlers[event] = make(map[command.Handler]struct{})
		}
		b.eventHandlers[event][handler] = struct{}{}
	}
}
