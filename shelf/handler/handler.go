package handler

import (
	"fmt"
	"github.com/segmentio/kafka-go"
	"mc-burger-orders/command"
	"mc-burger-orders/event"
	"mc-burger-orders/log"
	"mc-burger-orders/order"
	"mc-burger-orders/shelf"
	utils2 "mc-burger-orders/utils"
)

type Handler struct {
	defaultHandler command.DefaultCommandHandler
	KitchenService order.KitchenRequestService
	Shelf          *shelf.Shelf
}

func NewShelfHandler(kitchenTopicConfigs *event.TopicConfigs, s *shelf.Shelf) *Handler {
	return &Handler{
		Shelf:          s,
		KitchenService: NewKitchenService(kitchenTopicConfigs),
		defaultHandler: command.DefaultCommandHandler{},
	}
}

func (o *Handler) Handle(message kafka.Message) (bool, error) {
	commands, err := o.GetCommands(message)
	if err != nil {
		log.Error.Println(err.Error())
		return false, err
	}

	return o.defaultHandler.HandleCommands(message, commands...)
}

func (o *Handler) GetHandledEvents() []string {
	return []string{shelf.CheckFavoritesOnShelfEvent}
}

func (o *Handler) AddCommands(event string, commands ...command.Command) {
	o.defaultHandler.AddCommands(event, commands...)
}

func (o *Handler) GetCommands(message kafka.Message) ([]command.Command, error) {
	eventType, err := utils2.GetEventType(message)
	if err != nil {
		log.Error.Println(err.Error())
		return nil, err
	}

	commands := make([]command.Command, 0)
	switch eventType {
	case shelf.CheckFavoritesOnShelfEvent:
		{
			commands = append(commands, &RequestMissingItemsOnShelfCommand{
				Shelf:          o.Shelf,
				KitchenService: o.KitchenService,
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
