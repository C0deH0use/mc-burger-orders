package order

import (
	"context"
	"fmt"
	"github.com/segmentio/kafka-go"
	"mc-burger-orders/command"
	"mc-burger-orders/log"
	"net/http"
)

type OrderCollectedCommand struct {
	Repository    OrderRepository
	StatusEmitter StatusEmitter
	OrderNumber   int64
}

func (o *OrderCollectedCommand) Execute(ctx context.Context, message kafka.Message, commandResults chan command.TypedResult) {
	log.Info.Printf("Following order %d is going to be collected by the client", o.OrderNumber)
	order, err := o.Repository.FetchByOrderNumber(ctx, o.OrderNumber)
	if err != nil {
		errMessage := fmt.Sprintf("failed to find order by order number. Reason: %v", err.Error())
		commandResults <- command.NewHttpErrorResult("OrderCollectedCommand", errMessage, http.StatusNotFound)
		return
	}

	if isNotInRequiredStatus(order.Status) {
		errMessage := "requested order is yet ready for collection"
		commandResults <- command.NewHttpErrorResult("OrderCollectedCommand", errMessage, http.StatusPreconditionRequired)
		return
	}

	if order.Status == Collected {
		errMessage := "requested order already is collected"
		commandResults <- command.NewHttpErrorResult("OrderCollectedCommand", errMessage, http.StatusPreconditionFailed)
		return
	}

	order.Status = Collected
	log.Info.Printf("Order %d has been collected by customer", o.OrderNumber)
	_, err = o.Repository.InsertOrUpdate(ctx, order)
	if err != nil {
		errMessage := fmt.Sprintf("failed to update order `%d`, reason: %v", order.OrderNumber, err)
		commandResults <- command.NewHttpErrorResult("OrderCollectedCommand", errMessage, http.StatusInternalServerError)
		return
	}
	go o.StatusEmitter.EmitStatusUpdatedEvent(*order)
	commandResults <- command.NewSuccessfulResult("OrderCollectedCommand")
}

func isNotInRequiredStatus(status OrderStatus) bool {
	return status == Requested || status == InProgress
}
