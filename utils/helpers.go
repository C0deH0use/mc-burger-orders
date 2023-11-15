package utils

import (
	"fmt"
	"github.com/segmentio/kafka-go"
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
