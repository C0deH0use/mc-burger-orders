package order

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/segmentio/kafka-go"
	"mc-burger-orders/event"
	"mc-burger-orders/kitchen"
	"mc-burger-orders/order/dto"
	"mc-burger-orders/utils"
	"strconv"
	"time"
)

type KitchenRequestService interface {
	RequestNew(ctx context.Context, itemName string, quantity int) error
}

type KitchenService struct {
	*event.DefaultWriter
}

func NewKitchenServiceFrom(config *event.TopicConfigs) *KitchenService {
	defaultWriter := event.NewTopicWriter(config)
	return &KitchenService{defaultWriter}
}

func (s *KitchenService) RequestNew(ctx context.Context, itemName string, quantity int) error {
	headers := make([]kafka.Header, 0)
	headers = append(headers, utils.EventTypeHeader(kitchen.RequestItemEvent))

	message := make([]*dto.KitchenRequestMessage, 0)
	message = append(message, dto.NewKitchenRequestMessage(itemName, quantity))
	msgValue, err := json.Marshal(message)
	if err != nil {
		err = fmt.Errorf("failed to convert message details to bytes. Reason: %s", err)
		return err
	}

	msgKey := []byte(strconv.FormatInt(time.Now().UnixNano(), 10))
	msg := kafka.Message{
		Headers: headers,
		Key:     msgKey,
		Value:   msgValue,
	}
	if err = s.SendMessage(ctx, msg); err != nil {
		return err
	}
	return nil
}
