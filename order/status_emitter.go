package order

import (
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"mc-burger-orders/event"
	om "mc-burger-orders/order/model"
	"mc-burger-orders/utils"
	"strconv"
	"time"
)

type StatusEmitter interface {
	EmitStatusUpdatedEvent(order *om.Order)
}

type StatusEmitterService struct {
	OrderTopicConfig *event.TopicConfigs
}

func NewStatusEmitterFrom(topicConfig *event.TopicConfigs) *StatusEmitterService {
	return &StatusEmitterService{OrderTopicConfig: topicConfig}
}

func (r *StatusEmitterService) EmitStatusUpdatedEvent(order *om.Order) {
	writer := event.NewTopicWriter(r.OrderTopicConfig)

	headers := make([]kafka.Header, 0)
	headers = append(headers, utils.OrderHeader(order.OrderNumber))
	headers = append(headers, utils.EventTypeHeader(StatusUpdatedEvent))
	payloadBody := map[string]om.OrderStatus{
		"status": order.Status,
	}

	if payload, err := json.Marshal(payloadBody); err == nil {
		message := kafka.Message{
			Headers: headers,
			Key:     []byte(strconv.FormatInt(time.Now().UnixNano(), 10)),
			Value:   payload,
		}
		go func() {
			_ = writer.SendMessage(context.Background(), message)
		}()
	}
}