package order

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/segmentio/kafka-go"
	"mc-burger-orders/log"
)

type OrderUpdatedCommand struct {
	Repository  OrderRepository
	OrderNumber int64
}

func (o *OrderUpdatedCommand) Execute(ctx context.Context, message kafka.Message) (bool, error) {
	payloadBody := map[string]OrderStatus{}
	if err := json.Unmarshal(message.Value, &payloadBody); err != nil {
		log.Error.Printf("failed to unmarshal payload from message to map of strings. Reason: %v", err.Error())
		return false, err
	}

	status, exists := payloadBody["status"]
	if !exists {
		err := fmt.Errorf("failed to find status property in payload: +%v", payloadBody)
		return false, err
	}

	log.Info.Printf("Order %d has been updated to %v", o.OrderNumber, status)
	return true, nil
}
