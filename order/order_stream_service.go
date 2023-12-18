package order

import (
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"mc-burger-orders/event"
	"mc-burger-orders/utils"
	"strconv"
	"time"
)

type OrderStreamService interface {
	EmitUpdatedEvent(order *Order)
}

type OrderStreamServiceImpl struct {
	OrderTopicConfig *event.TopicConfigs
}

func NewOrderStreamService(topicConfig *event.TopicConfigs) OrderStreamService {
	return &OrderStreamServiceImpl{OrderTopicConfig: topicConfig}
}

func (r *OrderStreamServiceImpl) EmitUpdatedEvent(order *Order) {
	writer := event.NewTopicWriter(r.OrderTopicConfig)

	headers := make([]kafka.Header, 0)
	headers = append(headers, utils.OrderHeader(order.OrderNumber))
	headers = append(headers, utils.EventTypeHeader(OrderUpdatedEvent))

	if payload, err := json.Marshal(order); err == nil {
		message := kafka.Message{
			Headers: headers,
			Value:   payload,
			Key:     []byte(strconv.FormatInt(time.Now().UnixNano(), 10)),
		}
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			_ = writer.SendMessage(ctx, message)
			defer cancel()
		}()
	}
}
