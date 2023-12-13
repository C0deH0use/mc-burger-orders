package order

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/mongo"
	"mc-burger-orders/command"
	"mc-burger-orders/event"
	"mc-burger-orders/log"
	"mc-burger-orders/middleware"
	utils2 "mc-burger-orders/utils"
	"time"
)

type StatusEventDto struct {
	OrderNumber int64       `json:"orderNumber"`
	Status      OrderStatus `json:"status"`
}

type StatusEventsEndpoints struct {
	statusReader *event.DefaultReader
	repository   OrderRepository
	dispatcher   command.Dispatcher
}

func NewOrderStatusEventsEndpoints(database *mongo.Database, statusEmitterTopicConfigs *event.TopicConfigs) middleware.EndpointsSetup {
	repository := NewRepository(database)
	statusReader := event.NewTopicReader(statusEmitterTopicConfigs, nil)

	return &StatusEventsEndpoints{
		statusReader: statusReader,
		repository:   repository,
		dispatcher:   &command.DefaultDispatcher{},
	}
}

func (e *StatusEventsEndpoints) Setup(r *gin.Engine) {
	r.GET("/order/status", e.orderStatusHandler)
}

func (e *StatusEventsEndpoints) orderStatusHandler(c *gin.Context) {
	ctxWithTimeout, cancel := context.WithTimeout(c, time.Minute*3)
	topicChan := make(chan kafka.Message)

	go e.statusReader.SubscribeToTopic(topicChan)

	for eventMessage := range topicChan {
		eventUpdatePayload, err := e.validateMessage(ctxWithTimeout, eventMessage)
		if err != nil {
			log.Error.Println("Error when reading message from", e.statusReader.TopicName(), err.Error())
			_ = e.statusReader.Close()
			cancel()
			return
		}
		c.SSEvent("order-status", eventUpdatePayload)
		c.Writer.Flush()
	}

	defer func() {
		_ = e.statusReader.Close()
		close(topicChan)
		cancel()
	}()
}

func (e *StatusEventsEndpoints) validateMessage(ctx context.Context, message kafka.Message) (*StatusEventDto, error) {
	orderNumber, err := utils2.GetOrderNumber(message)

	if err != nil {
		return nil, err
	}

	_, err = e.repository.FetchByOrderNumber(ctx, orderNumber)
	if err != nil {
		err := fmt.Errorf("failed to find order by order number in payload: %v", orderNumber)
		return nil, err
	}
	payloadBody := map[string]OrderStatus{}
	if err := json.Unmarshal(message.Value, &payloadBody); err != nil {
		log.Error.Printf("failed to unmarshal payload from message to map of strings. Reason: %v", err.Error())
		return nil, err
	}

	status, exists := payloadBody["status"]
	if !exists {
		err := fmt.Errorf("failed to find status property in payload: +%v", payloadBody)
		return nil, err
	}

	return &StatusEventDto{OrderNumber: orderNumber, Status: status}, err
}
