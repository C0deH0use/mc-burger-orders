package event

import (
	"context"
	"fmt"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"mc-burger-orders/command"
	"mc-burger-orders/testing/utils"
	"strconv"
	"sync"
	"testing"
	"time"
)

var (
	kafkaConfig *TopicConfigs
	sut         *DefaultReader
	topic       = fmt.Sprintf("test-stack-updates-%d", rand.Intn(100))
	eventType   = "test-event"
)

type StubCommand struct {
	Invocations int
	waitG       *sync.WaitGroup
}

func (s *StubCommand) Execute(ctx context.Context, message kafka.Message, result chan command.TypedResult) {
	s.Invocations++
	s.waitG.Done()
	result <- command.TypedResult{Result: true, Type: "StubCommand"}
}

func (s *StubCommand) GetOrderNumber(message kafka.Message) (int64, error) {
	return int64(1010), nil
}

func TestIntegration_DefaultReader(t *testing.T) {
	utils.IntegrationTest(t)
	kafkaContainer, brokers := utils.TestWithKafka(t, context.Background())
	kafkaConfig = &TopicConfigs{
		Topic:             topic,
		Brokers:           brokers,
		NumPartitions:     1,
		ReplicationFactor: 1,
		AutoCreateTopic:   true,
	}

	t.Run("should consume new message send to topic", shouldConsumeNewMessageSendToTopic)

	t.Cleanup(func() {
		t.Log("Running Clean UP code")
		utils.TerminateKafka(t, context.Background(), kafkaContainer)
	})
}

func shouldConsumeNewMessageSendToTopic(t *testing.T) {
	// given
	stackMessages := make(chan kafka.Message)
	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(6)
	stubCommand := StubCommand{Invocations: 0, waitG: waitGroup}

	eventBus := NewInternalEventBus()
	commandHandler := command.NewCommandHandler()
	commandHandler.AddCommands(eventType, &stubCommand)

	eventBus.AddHandler(commandHandler)

	sut = NewTopicReader(kafkaConfig, eventBus)
	go sut.SubscribeToTopic(stackMessages)

	// when
	t.Log("Preparing to send test messages to topic", kafkaConfig.Topic)
	for msgId := range make([]int, 6) {
		go sendMessages(t, msgId)
	}

	// then
	waitGroup.Wait()

	assert.Equal(t, 6, stubCommand.Invocations, "all messages already processed!")
}

func sendMessages(t *testing.T, msgId int) {
	writer := NewTopicWriter(kafkaConfig)

	headers := make([]kafka.Header, 0)
	headers = append(headers, kafka.Header{Key: "order", Value: []byte(strconv.FormatInt(1010, 10))})
	headers = append(headers, kafka.Header{Key: "event", Value: []byte(eventType)})
	msgKey := []byte(strconv.FormatInt(time.Now().UnixNano(), 10))
	msgValue := fmt.Sprintf("Test Message %d", msgId)
	msg := kafka.Message{
		Key:     msgKey,
		Headers: headers,
		Value:   []byte(msgValue),
	}

	// when
	err := writer.SendMessage(context.Background(), msg)
	if err != nil {
		assert.Fail(t, "failed to send message to topic", kafkaConfig.Topic)
	}
}
