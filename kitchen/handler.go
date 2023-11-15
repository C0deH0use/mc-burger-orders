package kitchen

import (
	"github.com/gammazero/workerpool"
	"github.com/segmentio/kafka-go"
	"mc-burger-orders/command"
	"mc-burger-orders/event"
	"mc-burger-orders/log"
	"mc-burger-orders/stack"
	"mc-burger-orders/utils"
)

type Handler struct {
	command.DefaultDispatcher
	kitchenCooks        *workerpool.WorkerPool
	stack               *stack.Stack
	stackTopicConfig    *event.TopicConfigs
	kitchenTopicConfigs *event.TopicConfigs
}

func NewHandler(kitchenTopicConfigs *event.TopicConfigs, stackTopicConfig *event.TopicConfigs, s *stack.Stack) *Handler {
	return &Handler{
		kitchenCooks:        workerpool.New(1),
		stack:               s,
		stackTopicConfig:    stackTopicConfig,
		kitchenTopicConfigs: kitchenTopicConfigs,
	}
}

func (h *Handler) Handle(message kafka.Message) (bool, error) {

	eventType, err := utils.GetEventType(message)
	if err != nil {
		log.Error.Println(err.Error())
		return false, err
	}

	switch eventType {
	case RequestItemEvent:
		{
			h.kitchenCooks.Submit(func() {
				request := ItemRequest{ItemName: "item", Quantity: 1}
				_, err := h.CreateNewItem(request)

				if err != nil {
					log.Error.Println(err.Error())
				}
			})
		}
	}

	return true, nil
}

func (h *Handler) AwaitOn(msgChan chan kafka.Message) {
	select {
	case message := <-msgChan:
		{
			_, eventErr := h.Handle(message)
			if eventErr != nil {
				log.Error.Println(eventErr.Error())
			}
		}
	}
}
