package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/segmentio/kafka-go"
	"mc-burger-orders/event"
	"mc-burger-orders/kitchen"
	"mc-burger-orders/utils"
	"strconv"
	"time"
)

type NewItemRequestMessage struct {
	ItemName string `json:"itemName"`
	Quantity int    `json:"quantity"`
}

type KitchenService interface {
	RequestNew(ctx context.Context, itemName string, quantity int) error
}

type KitchenServiceImpl struct {
	*event.DefaultWriter
}

func NewKitchenService(config *event.TopicConfigs) *KitchenServiceImpl {
	defaultWriter := event.NewTopicWriter(config)
	return &KitchenServiceImpl{defaultWriter}
}

func (s *KitchenServiceImpl) RequestNew(ctx context.Context, itemName string, quantity int) error {
	headers := make([]kafka.Header, 0)
	headers = append(headers, utils.EventTypeHeader(kitchen.RequestItemEvent))

	message := make([]*NewItemRequestMessage, 0)
	message = append(message, &NewItemRequestMessage{itemName, quantity})
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
