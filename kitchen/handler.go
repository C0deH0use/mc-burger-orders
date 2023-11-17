package kitchen

import (
	"github.com/gammazero/workerpool"
	"github.com/segmentio/kafka-go"
	"github.com/spf13/cast"
	"mc-burger-orders/command"
	"mc-burger-orders/event"
	"mc-burger-orders/log"
	"mc-burger-orders/stack"
	"mc-burger-orders/utils"
	"os"
	"strconv"
)

type Handler struct {
	command.DefaultCommandHandler
	mealPreparation     MealPreparation
	kitchenCooks        *workerpool.WorkerPool
	stack               *stack.Stack
	stackMessageWriter  event.Writer
	kitchenTopicConfigs *event.TopicConfigs
}

func NewHandler(kitchenTopicConfigs *event.TopicConfigs, stackTopicConfig *event.TopicConfigs, s *stack.Stack) *Handler {
	maxWorkers := 3
	maxWorkersVal := os.Getenv("KITCHEN_WORKERS_MAX")

	if len(maxWorkersVal) > 0 {
		if value, err := strconv.ParseInt(maxWorkersVal, 10, 16); err == nil {
			maxWorkers = cast.ToInt(value)
		}
	}
	return &Handler{
		kitchenCooks:        workerpool.New(maxWorkers),
		mealPreparation:     &MealPreparationService{},
		stack:               s,
		stackMessageWriter:  event.NewTopicWriter(stackTopicConfig),
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
				_, err = h.CreateNewItem(message)

				if err != nil {
					log.Error.Println(err.Error())
				}
			})
		}
	}

	return true, nil
}
