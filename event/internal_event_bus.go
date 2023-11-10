package event

import (
	"github.com/segmentio/kafka-go"
	"mc-burger-orders/log"
	"reflect"
)

func typeOf(i interface{}) string {
	return reflect.TypeOf(i).Elem().Name()
}

type InternalEventBus struct {
	eventHandlers map[string]map[Handler]struct{}
}

func NewInternalEventBus() *InternalEventBus {
	return &InternalEventBus{
		eventHandlers: make(map[string]map[Handler]struct{}),
	}
}

// PublishEvent publishes events to all registered event handlers
func (b *InternalEventBus) PublishEvent(message kafka.Message) error {
	eventType := message.Topic
	if handlers, ok := b.eventHandlers[eventType]; ok {
		for handler := range handlers {
			result, err := handler.Handle(message)
			if err != nil {
				log.Error.Println("Error when processing following event:", eventType, "by handler -> ", reflect.TypeOf(handler))
				return err
			}
			log.Info.Println("Event", eventType, "was handled with result", result)
		}
	}
	return nil
}

// AddHandler registers an event handler for all the events specified in the
// variadic events parameter.
func (b *InternalEventBus) AddHandler(handler Handler, events ...interface{}) {

	for _, event := range events {
		typeName := typeOf(event)

		// There can be multiple handlers for any event.
		// Here we check that a map is initialized to hold these handlers
		// for a given type. If not we create one.
		if _, ok := b.eventHandlers[typeName]; !ok {
			b.eventHandlers[typeName] = make(map[Handler]struct{})
		}

		// Add this handler to the collection of handlers for the type.
		b.eventHandlers[typeName][handler] = struct{}{}
	}
}
