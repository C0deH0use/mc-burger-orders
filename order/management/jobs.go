package management

import (
	"context"
	"github.com/segmentio/kafka-go"
	"mc-burger-orders/event"
	"mc-burger-orders/log"
	utils2 "mc-burger-orders/utils"
	"strconv"
	"time"
)

func OrderManagementJobs() {
	topicConfigs := OrderJobsTopicConfigsFromEnv()
	writer := event.NewTopicWriter(topicConfigs)
	for {
		go func() {
			msg := CheckOrdersMissingItemsMessage()
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			err := writer.SendMessage(ctx, msg)
			if err != nil {
				log.Error.Printf("failed to publish message on topic %v. Error reason: %v", topicConfigs.Topic, err)
			}
			defer cancel()
		}()

		time.Sleep(1 * time.Minute)
	}
}

func CheckOrdersMissingItemsMessage() kafka.Message {
	headers := make([]kafka.Header, 0)
	headers = append(headers, utils2.EventTypeHeader(CheckMissingItemsOnOrdersEvent))
	msgKey := []byte(strconv.FormatInt(time.Now().UnixNano(), 10))

	return kafka.Message{
		Headers: headers,
		Key:     msgKey,
	}
}
