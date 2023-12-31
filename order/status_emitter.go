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

type StatusEmitter interface {
	EmitStatusUpdatedEvent(order Order)
}

type StatusEmitterService struct {
	OrderTopicConfig *event.TopicConfigs
}

func NewStatusEmitterFrom(topicConfig *event.TopicConfigs) *StatusEmitterService {
	return &StatusEmitterService{OrderTopicConfig: topicConfig}
}

func (r *StatusEmitterService) EmitStatusUpdatedEvent(o Order) {
	order := o
	writer := event.NewTopicWriter(r.OrderTopicConfig)

	headers := make([]kafka.Header, 0)
	headers = append(headers, utils.OrderHeader(order.OrderNumber))
	headers = append(headers, utils.EventTypeHeader(StatusUpdatedEvent))
	payloadBody := map[string]OrderStatus{
		"status": order.Status,
	}

	if payload, err := json.Marshal(payloadBody); err == nil {
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
