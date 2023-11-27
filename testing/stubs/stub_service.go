package stubs

import (
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"log"
	m "mc-burger-orders/order"
)

type StubService struct {
	o            *m.Order
	err          error
	methodCalled []map[string]interface{}
}

func NewStubService() *StubService {
	return &StubService{methodCalled: make([]map[string]any, 0)}
}

func NewStubServiceWithOrder(order *m.Order) *StubService {
	return &StubService{o: order, err: nil, methodCalled: make([]map[string]any, 0)}
}

func NewStubServiceWithErr(err error) *StubService {
	return &StubService{err: err, methodCalled: make([]map[string]any, 0)}
}

func (s *StubService) CalledCnt() int {
	return len(s.methodCalled)
}

func RequestMatchingFnc(itemName string, quantity int, orderNumber int64) func(args map[string]any) bool {
	return func(args map[string]any) bool {
		argName := args["itemName"]
		argQuantity := args["quantity"]
		argNumber := args["orderNumber"]
		log.Printf("StubService methodCalled. %+v", args)
		b := argName == itemName && argQuantity == quantity && argNumber == orderNumber
		return b
	}
}

func MealPrepMatchingFnc(itemName string, quantity int) func(args map[string]any) bool {
	return func(args map[string]any) bool {
		argName := args["itemName"]
		argQuantity := args["quantity"]
		log.Printf("StubService methodCalled. %+v", args)
		return argName == itemName && argQuantity == quantity
	}
}

func StatusUpdateMatchingFnc(status m.OrderStatus) func(args map[string]any) bool {
	return func(args map[string]any) bool {
		if statusUpdated, exists := args["StatusUpdatedEvent"]; exists {
			log.Printf("StubService methodCalled. %+v", args)
			return statusUpdated == status
		}
		return false
	}
}

func KafkaMessageMatchingFnc(orderNumber int64, messageVal []map[string]any) func(args map[string]any) bool {
	if b, err := json.Marshal(messageVal); err == nil {

		expectedMessageValue := string(b)
		return func(args map[string]any) bool {
			if sendMessageArg, ok := args["SendMessage"]; ok {
				messages := sendMessageArg.([]kafka.Message)

				for _, kafkaMsg := range messages {
					number, err := m.GetOrderNumber(kafkaMsg)
					if orderNumber == -1 && err != nil {
						if string(kafkaMsg.Value) == expectedMessageValue {
							return true
						}
					}
					if orderNumber > 0 && err != nil {
						return false
					}

					if number == orderNumber && string(kafkaMsg.Value) == expectedMessageValue {
						return true
					}
				}
				return false
			}
			return false
		}
	}
	return func(args map[string]any) bool {
		return false
	}
}

func (s *StubService) HaveBeenCalledWith(matchingFnc func(args map[string]any) bool) bool {
	r := false

	for _, args := range s.methodCalled {
		if matchingFnc(args) {
			r = true
		}
	}

	return r
}

func (s *StubService) RequestForOrder(ctx context.Context, itemName string, quantity int, orderNumber int64) error {
	args := map[string]interface{}{
		"itemName":    itemName,
		"quantity":    quantity,
		"orderNumber": orderNumber,
	}
	s.methodCalled = append(s.methodCalled, args)
	return nil
}

func (s *StubService) SendMessage(ctx context.Context, messages ...kafka.Message) error {
	args := map[string]interface{}{
		"SendMessage": messages,
	}
	s.methodCalled = append(s.methodCalled, args)
	return nil
}

func (s *StubService) Prepare(item string, quantity int) {
	args := map[string]interface{}{
		"itemName": item,
		"quantity": quantity,
	}
	s.methodCalled = append(s.methodCalled, args)
}

func (s *StubService) EmitStatusUpdatedEvent(order *m.Order) {
	args := map[string]interface{}{
		"StatusUpdatedEvent": order.Status,
	}
	s.methodCalled = append(s.methodCalled, args)
}
