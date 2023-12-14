package order

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/segmentio/kafka-go"
	"mc-burger-orders/command"
	"mc-burger-orders/log"
)

type OrderUpdatedCommand struct {
	Repository  OrderRepository
	OrderNumber int64
}

func (o *OrderUpdatedCommand) Execute(ctx context.Context, message kafka.Message, commandResults chan command.TypedResult) {
	payloadBody := map[string]OrderStatus{}
	if err := json.Unmarshal(message.Value, &payloadBody); err != nil {
		log.Error.Printf("failed to unmarshal payload from message to map of strings. Reason: %v", err.Error())
		commandResults <- command.NewErrorResult("OrderUpdatedCommand", err)
		return
	}
	_, err := o.Repository.FetchByOrderNumber(ctx, o.OrderNumber)
	if err != nil {
		err := fmt.Errorf("failed to find order by order number in payload: %v", o.OrderNumber)
		commandResults <- command.NewErrorResult("OrderUpdatedCommand", err)
		return
	}

	status, exists := payloadBody["status"]
	if !exists {
		err := fmt.Errorf("failed to find status property in payload: +%v", payloadBody)
		commandResults <- command.NewErrorResult("OrderUpdatedCommand", err)
		return
	}

	log.Info.Printf("Order %d has been updated to %v", o.OrderNumber, status)
	commandResults <- command.NewSuccessfulResult("OrderUpdatedCommand")
}
