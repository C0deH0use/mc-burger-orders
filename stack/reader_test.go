package stack

import (
	"context"
	"fmt"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"log"
	"math/rand"
	"mc-burger-orders/command"
	"mc-burger-orders/event"
	"mc-burger-orders/utils"
	"strconv"
	"sync"
	"testing"
	"time"
)

var (
	kafkaConfig *event.TopicConfigs
	sut         *event.DefaultReader
	topic       = fmt.Sprintf("test-stack-updates-%d", rand.Intn(100))
)

type StubCommand struct {
	Invocations int
	waitG       *sync.WaitGroup
}

func (s *StubCommand) Execute(ctx context.Context) (bool, error) {
	s.Invocations++
	s.waitG.Done()
	return true, nil
}

func TestStackReader(t *testing.T) {
	ctx := context.Background()
	kafkaContainer, brokers := utils.TestWithKafka(ctx)
	kafkaConfig = &event.TopicConfigs{
		Topic:             topic,
		Brokers:           brokers,
		NumPartitions:     1,
		ReplicationFactor: 1,
		AutoCreateTopic:   true,
	}

	t.Run("should consume new message send to topic", shouldConsumeNewMessageSendToTopic)

	t.Cleanup(func() {
		log.Println("Running Clean UP code")
		utils.TerminateKafka(kafkaContainer)
	})
}

func shouldConsumeNewMessageSendToTopic(t *testing.T) {
	// given
	stackMessages := make(chan kafka.Message)
	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(6)
	stubCommand := StubCommand{Invocations: 0, waitG: waitGroup}

	eventBus := event.NewInternalEventBus()
	commandHandler := command.NewCommandHandler()
	commandHandler.AddCommands(topic, &stubCommand)

	eventBus.AddHandler(commandHandler, topic)

	sut = event.NewTopicReader(kafkaConfig, eventBus)
	sut.SubscribeToTopic(stackMessages)

	// when
	log.Println("Preparing to send test messages to topic", kafkaConfig.Topic)
	for _, msgId := range []int{1, 2, 3, 4, 5, 6} {
		go sendMessages(t, msgId)
	}

	// then
	waitGroup.Wait()

	assert.Equal(t, 6, stubCommand.Invocations, "all messages already processed!")
}

func sendMessages(t *testing.T, msgId int) {
	writer := event.NewTopicWriter(kafkaConfig)

	msgKey := []byte(strconv.FormatInt(time.Now().UnixNano(), 10))
	msgValue := fmt.Sprintf("Test Message %d", msgId)
	msg := kafka.Message{
		Key:   msgKey,
		Value: []byte(msgValue),
	}

	// when
	log.Printf("Sending %d-th message on topic %v", msgId, kafkaConfig.Topic)

	s, err := time.ParseDuration(strconv.FormatInt(rand.Int63n(500), 10))
	var randSleep = time.Millisecond * s

	time.Sleep(randSleep)
	err = writer.SendMessage(context.Background(), msg)
	if err != nil {
		assert.Fail(t, "failed to send message to topic", kafkaConfig.Topic)
	}
}
