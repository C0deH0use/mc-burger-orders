package utils

import (
	"fmt"
	"github.com/segmentio/kafka-go"
	"strconv"
)

func GetEventType(message kafka.Message) (string, error) {
	for _, header := range message.Headers {
		if header.Key == "event" {
			return string(header.Value), nil
		}
	}
	err := fmt.Errorf("count not find event header in messgae")
	return "", err
}

func OrderHeader(orderNumber int64) kafka.Header {
	return kafka.Header{Key: "order", Value: []byte(strconv.FormatInt(orderNumber, 10))}
}

func EventTypeHeader(eventType string) kafka.Header {
	return kafka.Header{Key: "event", Value: []byte(eventType)}
}
