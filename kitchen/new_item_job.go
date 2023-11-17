package kitchen

import (
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"mc-burger-orders/log"
	"mc-burger-orders/order/utils"
	stack2 "mc-burger-orders/stack"
	utils2 "mc-burger-orders/utils"
	"strconv"
	"time"
)

func (h *Handler) CreateNewItem(message kafka.Message) (bool, error) {
	requests := &[]ItemRequest{}
	err := json.Unmarshal(message.Value, requests)
	if err != nil {
		log.Error.Println(err.Error())
		return false, err
	}
	orderNumber, err := utils.GetOrderNumber(message)
	if err != nil {
		orderNumber = -1
	}

	messageKey := string(message.Key)
	messages := make([]kafka.Message, 0)
	log.Info.Printf("CookRequest: %v | Order: %d", messageKey, orderNumber)

	for _, request := range *requests {
		log.Info.Printf("CookRequest: %v | Order: %d | Starting to prepare new item -> %v in amount: `%d`", messageKey, orderNumber, request.ItemName, request.Quantity)

		h.mealPreparation.Prepare(request.ItemName, request.Quantity)

		log.Info.Printf("CookRequest: %v | Order: %d | item(s) %v in prepared", messageKey, orderNumber, request.ItemName)

		if newMessage, err := buildMessage(orderNumber, request); err == nil {
			h.stack.AddMany(request.ItemName, request.Quantity)
			messages = append(messages, newMessage)
		}
	}

	if len(messages) > 0 {
		err = h.stackMessageWriter.SendMessage(context.Background(), messages...)
		if err != nil {
			log.Error.Println(err.Error())
			return false, err
		}
	}

	return true, nil
}

func buildMessage(orderNumber int64, item ItemRequest) (kafka.Message, error) {
	headers := make([]kafka.Header, 0)
	headers = append(headers, utils2.OrderHeader(orderNumber))
	headers = append(headers, utils2.EventTypeHeader(stack2.ItemAddedToStackEvent))

	b, err := json.Marshal(item)
	if err != nil {
		return kafka.Message{}, err
	}

	return kafka.Message{
		Headers: headers,
		Key:     []byte(strconv.FormatInt(time.Now().UnixNano(), 10)),
		Value:   b,
	}, nil
}
