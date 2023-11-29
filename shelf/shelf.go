package shelf

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/segmentio/kafka-go"
	"github.com/spf13/cast"
	"mc-burger-orders/event"
	"mc-burger-orders/kitchen/item"
	"mc-burger-orders/log"
	"mc-burger-orders/shelf/dto"
	utils2 "mc-burger-orders/utils"
	"strconv"
	"sync"
	"time"
)

type Shelf struct {
	writer *event.DefaultWriter
	data   *sync.Map
}

func NewEmptyShelf() *Shelf {
	return &Shelf{data: CleanShelf()}
}

func CleanShelf() *sync.Map {
	syncMap := &sync.Map{}
	for itemConfig := range item.MenuItems {
		if !item.MenuItems[itemConfig].InstantReady {
			syncMap.Store(itemConfig, 0)
		}
	}

	return syncMap
}

func (s *Shelf) ConfigureWriter(writer *event.DefaultWriter) {
	s.writer = writer
}

func (s *Shelf) Add(item string) {
	if value, ok := s.data.Load(item); ok {
		newVal := cast.ToInt(value) + 1
		s.data.Store(item, newVal)
		log.Warning.Printf("Kitchen Shelf | %v +1 => %d", item, newVal)
	} else {
		s.data.Store(item, 1)
		log.Warning.Printf("Kitchen Shelf | %v +1 => %d", item, 1)
	}

	s.SendUpdateEvent(item, 1)
}

func (s *Shelf) AddMany(item string, quantity int) {
	if value, ok := s.data.Load(item); ok {
		newVal := cast.ToInt(value) + quantity
		s.data.Store(item, newVal)
		log.Warning.Printf("Kitchen Shelf | %v + %d => %d", item, quantity, newVal)
	} else {
		s.data.Store(item, quantity)
		log.Warning.Printf("Kitchen Shelf | %v + %d=> %d", item, quantity, quantity)
	}

	s.SendUpdateEvent(item, quantity)
}

func (s *Shelf) GetCurrent(item string) int {
	if value, ok := s.data.Load(item); ok {
		return cast.ToInt(value)
	}
	return 0
}

func (s *Shelf) Take(item string, quantity int) error {
	if value, ok := s.data.Load(item); ok {
		exists := cast.ToInt(value)
		if exists < quantity {
			err := fmt.Errorf("not enought in Shelf of type %quantity. Requered %d, but only %d available", item, exists, quantity)
			return err
		}

		newVal := exists - quantity
		s.data.Store(item, newVal)
		log.Warning.Printf("Kitchen Shelf | %v - %d => %d", item, quantity, newVal)
		return nil
	}
	err := fmt.Errorf("unknown item `%v` requested", item)
	return err
}

func (s *Shelf) SendUpdateEvent(item string, quantity int) {
	if s.writer == nil {
		log.Warning.Printf("Shelf Events emitter not configured yet!")
		return
	}

	if kafkaMessage, err := createMessage(item, quantity); err == nil {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		if err = s.writer.SendMessage(ctx, kafkaMessage); err != nil {
			log.Error.Println("failed to send message to topic", s.writer.TopicName())
		}

		defer cancel()
	}
}

func createMessage(itemName string, quantity int) (kafka.Message, error) {
	headers := make([]kafka.Header, 0)
	headers = append(headers, utils2.EventTypeHeader(ItemAddedOnShelfEvent))
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
