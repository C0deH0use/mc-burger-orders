package event

import (
	"github.com/segmentio/kafka-go"
	"mc-burger-orders/command"
	"mc-burger-orders/log"
	"mc-burger-orders/utils"
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

func (b *InternalEventBus) PublishEvent(message kafka.Message, resultsCommands chan command.TypedResult) {
	eventType, err := utils.GetEventType(message)
	if err != nil {
		resultsCommands <- command.NewErrorResult(eventType, err)
		close(resultsCommands)
		return
	}

	if handlers, ok := b.eventHandlers[eventType]; ok {
		for handler := range handlers {
			handler.Handle(message, resultsCommands)
		}
		return
	}
	log.Error.Printf("Event %v does not have any handled", eventType)
}

func (b *InternalEventBus) AddHandler(handler command.Handler) {

	for _, event := range handler.GetHandledEvents() {

		if _, ok := b.eventHandlers[event]; !ok {
			b.eventHandlers[event] = make(map[command.Handler]struct{})
		}
		b.eventHandlers[event][handler] = struct{}{}
	}
}
