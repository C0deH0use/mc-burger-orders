package schedule

import (
	"context"
	"github.com/segmentio/kafka-go"
	"mc-burger-orders/event"
	"mc-burger-orders/log"
	"mc-burger-orders/shelf"
	utils2 "mc-burger-orders/utils"
	"strconv"
	"time"
)

func ShelfJobs() {
	topicConfigs := shelf.TopicConfigsFromEnv()
	writer := event.NewTopicWriter(topicConfigs)
	for {
		go func() {
			msg := CheckFavoritesOnShelfMessage()
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			err := writer.SendMessage(ctx, msg)
			if err != nil {
				log.Error.Printf("failed to publish message on topic %v. Error reason: %v", topicConfigs.Topic, err)
			}
			defer cancel()
		}()

		time.Sleep(3 * time.Minute)
	}
}

func CheckFavoritesOnShelfMessage() kafka.Message {
	headers := make([]kafka.Header, 0)
	headers = append(headers, utils2.EventTypeHeader(shelf.CheckFavoritesOnShelfEvent))
	msgKey := []byte(strconv.FormatInt(time.Now().UnixNano(), 10))

	return kafka.Message{
		Headers: headers,
		Key:     msgKey,
	}
}
