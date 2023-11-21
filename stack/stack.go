package stack

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/segmentio/kafka-go"
	"github.com/spf13/cast"
	"mc-burger-orders/event"
	"mc-burger-orders/kitchen/item"
	"mc-burger-orders/log"
	"mc-burger-orders/stack/dto"
	utils2 "mc-burger-orders/utils"
	"strconv"
	"sync"
	"time"
)

type Stack struct {
	writer       *event.DefaultWriter
	kitchenStack *sync.Map
}

func NewStack(ks *sync.Map) *Stack {
	return &Stack{kitchenStack: ks}
}

func NewEmptyStack() *Stack {
	return &Stack{kitchenStack: CleanStack()}
}

func CleanStack() *sync.Map {
	syncMap := &sync.Map{}
	for itemConfig := range item.MenuItems {
		if !item.MenuItems[itemConfig].InstantReady {
			syncMap.Store(itemConfig, 0)
		}
	}

	return syncMap
}

func (s *Stack) ConfigureWriter(writer *event.DefaultWriter) {
	s.writer = writer
}

func (s *Stack) Add(item string) {
	if value, ok := s.kitchenStack.Load(item); ok {
		newVal := cast.ToInt(value) + 1
		s.kitchenStack.Store(item, newVal)
		log.Warning.Printf("Kitchen Stack | %v => %d", item, newVal)
	} else {
		s.kitchenStack.Store(item, 1)
		log.Warning.Printf("Kitchen Stack | %v => + %d", item, 1)
	}

	go s.sendStackUpdateEvent(item, 1)
}

func (s *Stack) AddMany(item string, quantity int) {
	if value, ok := s.kitchenStack.Load(item); ok {
		newVal := cast.ToInt(value) + quantity
		s.kitchenStack.Store(item, newVal)
		log.Warning.Printf("Kitchen Stack | %v => %d", item, newVal)
	} else {
		s.kitchenStack.Store(item, quantity)
		log.Warning.Printf("Kitchen Stack | %v => %d", item, quantity)
	}

	go s.sendStackUpdateEvent(item, quantity)
}

func (s *Stack) GetCurrent(item string) int {
	if value, ok := s.kitchenStack.Load(item); ok {
		return cast.ToInt(value)
	}
	return 0
}

func (s *Stack) Take(item string, quantity int) error {
	if value, ok := s.kitchenStack.Load(item); ok {
		exists := cast.ToInt(value)
		if exists < quantity {
			err := fmt.Errorf("not enought in Stack of type %quantity. Requered %d, but only %d available", item, exists, quantity)
			return err
		}

		newVal := exists - quantity
		s.kitchenStack.Store(item, newVal)
		log.Warning.Printf("Kitchen Stack | %v => %d", item, newVal)
		return nil
	}
	err := fmt.Errorf("unknown item `%v` requested", item)
	return err
}

func (s *Stack) sendStackUpdateEvent(item string, quantity int) {
	if s.writer == nil {
		log.Warning.Printf("Stack Events emitter not configured yet!")
		return
	}

	if kafkaMessage, err := stackUpdatedMessage(item, quantity); err == nil {
		err = s.writer.SendMessage(context.Background(), kafkaMessage)
		if err != nil {
			log.Error.Println("failed to send message to topic", s.writer.TopicName())
		}
	}
}

func stackUpdatedMessage(itemName string, quantity int) (kafka.Message, error) {
	headers := make([]kafka.Header, 0)
	headers = append(headers, utils2.EventTypeHeader(ItemAddedToStackEvent))
	items := make([]dto.ItemAdded, 0)
	items = append(items, dto.ItemAdded{ItemName: itemName, Quantity: quantity})
	b, err := json.Marshal(items)

	if err != nil {
		return kafka.Message{}, err
	}

	return kafka.Message{
		Headers: headers,
		Key:     []byte(strconv.FormatInt(time.Now().UnixNano(), 10)),
		Value:   b,
	}, nil
}
