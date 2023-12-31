package event

import (
	"context"
	"fmt"
	"github.com/segmentio/kafka-go"
	"mc-burger-orders/command"
	"mc-burger-orders/log"
	"mc-burger-orders/utils"
	"time"
)

type NewMessageHandler interface {
	OnNewMessage(message kafka.Message) error
	HandleError(err error, message kafka.Message)
}

type DefaultReader struct {
	*kafka.Reader
	configuration *TopicConfigs
	eventBus      EventBus
	processRepeat map[string]int
}

func NewTopicReader(configuration *TopicConfigs, eventBus EventBus) *DefaultReader {
	log.Warning.Printf("Creating a new topicReader for topic: %v", configuration.Topic)
	if len(configuration.Brokers) == 0 {
		log.Error.Panicln("missing at least one Kafka Address")
	}

	conn := configuration.ConnectToBroker()
	configuration.CreateTopic(conn)

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  configuration.Brokers,
		Topic:    configuration.Topic,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
		MaxWait:  configuration.WaitMaxTime,
		GroupID:  groupID(configuration),
	})
	return &DefaultReader{reader, configuration, eventBus, make(map[string]int)}
}

func (r *DefaultReader) GroupId() string {
	return groupID(r.configuration)
}

func groupID(configuration *TopicConfigs) string {
	return fmt.Sprintf("%v-%d", configuration.Topic, configuration.Partition)
}

func (r *DefaultReader) TopicName() string {
	return r.configuration.Topic
}

func (r *DefaultReader) SubscribeToTopic(msgChan chan kafka.Message) {
	go func() {
		log.Info.Println("Subscribing to topic", r.configuration.Topic)

		for {
			r.ReadMessageFromTopic(context.Background(), msgChan)
			time.Sleep(r.configuration.AwaitBetweenReadsTime)
		}
	}()

	go func() {
		if r.eventBus != nil {
			for newMessage := range msgChan {
				go r.PublishEvent(newMessage)
			}
		}
	}()
}

func (r *DefaultReader) Close() error {
	return r.Reader.Close()
}

func (r *DefaultReader) PublishEvent(message kafka.Message) {
	// TODO: is Message Already Ran
	commandResults := make(chan command.TypedResult)
	r.eventBus.PublishEvent(message, commandResults)

	for commandResult := range commandResults {
		if commandResult.Error != nil {
			log.Error.Println("While executing command", commandResult.Type, "following error occurred", commandResult.Error.Error())
			r.HandleError(commandResult.Error.Error(), message)
		} else {
			log.Info.Println("Command", commandResult.Type, "finished successfully, with result -", commandResult.Result)
		}
	}
}

func (r *DefaultReader) ReadMessageFromTopic(ctx context.Context, msgChan chan kafka.Message) {
	msg, err := r.ReadMessage(ctx)
	if err != nil {
		log.Error.Println("failed to read message from topic:", r.configuration.Topic, "GroupID", r.GroupId(), err)
		return
	}
	eventType, err := utils.GetEventType(msg)
	if err != nil {
		log.Error.Println("failed to read event type from message:", err)
		return
	}

	log.Warning.Printf("Received messaged for topic: %v, event: %v", msg.Topic, eventType)
	if msg.Topic == r.configuration.Topic {
		msgChan <- msg
	}
}

func (r *DefaultReader) HandleError(err error, message kafka.Message) {
	key := string(message.Key)
	topic := message.Topic
	log.Error.Printf("failed to process message [%v] from topic: %v on event bus: %v\n", key, topic, err.Error())

	repeatCnt, ok := r.processRepeat[key]
	if !ok {
		repeatCnt = 0
	}

	if repeatCnt < 4 {
		// TODO: Error handling -> repeat message with delay
		log.Error.Printf("Message [%v] is going to be send back to topic to be repeater in processing\n", key)
		r.processRepeat[key] = repeatCnt + 1
		return
	}

	if repeatCnt >= 4 {
		log.Error.Printf("Message [%v] was already repeated %d. Going to be send message to dead-letter queue\n", key, repeatCnt)
		return
	}

	// default error handling
	log.Error.Println("default error handling when event could not be processed", err.Error())
}
