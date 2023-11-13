package event

import (
	"github.com/segmentio/kafka-go"
	"mc-burger-orders/command"
	"mc-burger-orders/log"
	"reflect"
)

type InternalEventBus struct {
	eventHandlers map[string]map[command.Handler]struct{}
}

func NewInternalEventBus() *InternalEventBus {
	return &InternalEventBus{
		eventHandlers: make(map[string]map[command.Handler]struct{}),
	}
}

// PublishEvent publishes events to all registered event handlers
func (b *InternalEventBus) PublishEvent(message kafka.Message) error {
	eventType := message.Topic
	if handlers, ok := b.eventHandlers[eventType]; ok {
		for handler := range handlers {
			result, err := handler.Handle(message)
			if err != nil {
				log.Error.Printf("Error when processing following event: %v by handler -> %v\n", eventType, reflect.TypeOf(handler))
				return err
			}
			log.Info.Printf("Event %v was handled with result %v\n", eventType, result)
		}
	}
	return nil
}

// AddHandler registers an event handler for all the events specified in the
// variadic events parameter.
func (b *InternalEventBus) AddHandler(handler command.Handler, events ...string) {

	for _, event := range events {

		// There can be multiple handlers for any event.
		// Here we check that a map is initialized to hold these handlers
		// for a given type. If not we create one.
		if _, ok := b.eventHandlers[event]; !ok {
			b.eventHandlers[event] = make(map[command.Handler]struct{})
		}

		// Add this handler to the collection of handlers for the type.
		b.eventHandlers[event][handler] = struct{}{}
	}
}
