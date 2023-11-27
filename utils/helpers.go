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

func GetOrderNumber(message kafka.Message) (int64, error) {
	var err error
	var orderNumber int64 = -1

	for _, header := range message.Headers {
		if header.Key == "order" {
			orderNumberStr := string(header.Value)

			orderNumber, err = strconv.ParseInt(orderNumberStr, 10, 64)
			if err != nil {
				err := fmt.Errorf("cannot parse order number '%v' to int64. %v", orderNumberStr, err)
				return -1, err
			}
		}
	}

	if orderNumber < 0 {
		err := fmt.Errorf("cannot find order number in message headers")
		return -1, err
	}

	return orderNumber, nil
}

func OrderHeader(orderNumber int64) kafka.Header {
	return kafka.Header{Key: "order", Value: []byte(strconv.FormatInt(orderNumber, 10))}
}

func EventTypeHeader(eventType string) kafka.Header {
	return kafka.Header{Key: "event", Value: []byte(eventType)}
}
