package kitchen

import (
	"github.com/gammazero/workerpool"
	"github.com/segmentio/kafka-go"
	"github.com/spf13/cast"
	"mc-burger-orders/command"
	"mc-burger-orders/log"
	"mc-burger-orders/shelf"
	"mc-burger-orders/utils"
	"os"
	"strconv"
)

type Handler struct {
	defaultHandler  command.DefaultCommandHandler
	mealPreparation MealPreparation
	kitchenCooks    *workerpool.WorkerPool
	shelf           *shelf.Shelf
}

func NewHandler(s *shelf.Shelf) *Handler {
	maxWorkers := 5
	maxWorkersVal := os.Getenv("KITCHEN_WORKERS_MAX")

	if len(maxWorkersVal) > 0 {
		if value, err := strconv.ParseInt(maxWorkersVal, 10, 16); err == nil {
			maxWorkers = cast.ToInt(value)
		}
	}
	return &Handler{
		kitchenCooks:    workerpool.New(maxWorkers),
		mealPreparation: &MealPreparationService{},
		shelf:           s,
		defaultHandler:  command.DefaultCommandHandler{},
	}
}

func (h *Handler) GetHandledEvents() []string {
	return []string{RequestItemEvent}
}

func (h *Handler) AddCommands(event string, commands ...command.Command) {
	h.defaultHandler.AddCommands(event, commands...)
}

func (h *Handler) GetCommands(_ kafka.Message) ([]command.Command, error) {
	return make([]command.Command, 0), nil
}

func (h *Handler) Handle(message kafka.Message, commandResults chan command.TypedResult) {
	eventType, err := utils.GetEventType(message)
	if err != nil {
		log.Error.Println(err.Error())
		commandResults <- command.NewErrorResult("GetCommandForMessage", err)
		return
	}

	switch eventType {
	case RequestItemEvent:
		{
			h.kitchenCooks.Submit(func() {
				_, err = h.CreateNewItem(message)

				if err != nil {
					log.Error.Println(err.Error())
					commandResults <- command.NewErrorResult(RequestItemEvent, err)
				} else {
					commandResults <- command.NewSuccessfulResult(RequestItemEvent)
				}
			})
		}
	}
}
