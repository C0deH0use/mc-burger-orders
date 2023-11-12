package stack

import (
	"context"
	"fmt"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"log"
	"math/rand"
	"mc-burger-orders/event"
	"mc-burger-orders/utils"
	"strconv"
	"testing"
	"time"
)

var (
	kafkaConfig *event.TopicConfigs
	sut         *event.DefaultReader
	topic       = fmt.Sprintf("test-stack-updates-%d", rand.Intn(100))
)

func TestStackReader(t *testing.T) {
	ctx := context.Background()
	kafkaContainer := utils.TestWithKafka(ctx)
	brokers, err := kafkaContainer.Brokers(ctx)
	if err != nil {
		assert.Fail(t, "cannot read Brokers from kafka container")
	}
	kafkaConfig = &event.TopicConfigs{
		Topic:             topic,
		Brokers:           brokers,
		NumPartitions:     1,
		ReplicationFactor: 1,
		AutoCreateTopic:   true,
	}
	sut = event.NewTopicReader(kafkaConfig)

	t.Run("should consume new message send to topic", shouldConsumeNewMessageSendToTopic)

	t.Cleanup(func() {
		log.Println("Running Clean UP code")
		utils.TerminateKafka(kafkaContainer)
	})
}

func shouldConsumeNewMessageSendToTopic(t *testing.T) {
	// given
	messages := make(chan kafka.Message)
	writer := event.NewTopicWriter(kafkaConfig)
	calls := []int{1, 2, 3, 4, 5, 6}

	// when
	go func() {
		log.Println("Preparing to send test messages to topic", kafkaConfig.Topic)

		time.Sleep(100 * time.Millisecond)

		for idx := range calls {
			msgKey := []byte(strconv.FormatInt(time.Now().UnixNano(), 10))
			msgValue := fmt.Sprintf("Test Message %d", idx)
			msg := kafka.Message{
				Key:   msgKey,
				Value: []byte(msgValue),
			}

			// when
			log.Println("Sending", idx, "message on topic", kafkaConfig.Topic)
			err := writer.SendMessage(context.Background(), msg)
			if err != nil {
				assert.Fail(t, "failed to send message to topic", kafkaConfig.Topic)
			}
		}
	}()
	// then
	go func() {
		sut.SubscribeToTopic(context.Background(), messages)
	}()

	msgCnt := 0
	for {
		select {
		case message := <-messages:
			{
				messageKey := string(message.Key)
				messageValue := string(message.Value)
				log.Println("New Stack message read, key: ", messageKey, "message value:", messageValue)
				msgCnt++
				if msgCnt == len(calls) {
					assert.Equal(t, len(calls), msgCnt)
					return
				}
			}
		case <-time.After(15 * time.Second):
			assert.Equal(t, 6, msgCnt, "read all messages")
		}
	}

}
