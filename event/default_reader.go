package event

import (
	"context"
	"github.com/segmentio/kafka-go"
	"mc-burger-orders/log"
)

type DefaultReader struct {
	*kafka.Reader
	configuration *TopicConfigs
}

func NewTopicReader(configuration *TopicConfigs) *DefaultReader {
	if len(configuration.Brokers) == 0 {
		log.Error.Panicln("missing at least one Kafka Address")
	}

	conn := configuration.ConnectToBroker()
	configuration.CreateTopic(conn)

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   configuration.Brokers,
		Topic:     configuration.Topic,
		Partition: 0,
		MinBytes:  10e3, // 10KB
		MaxBytes:  10e6, // 10MB
	})
	return &DefaultReader{reader, configuration}
}

func (r *DefaultReader) SubscribeToTopic(ctx context.Context, msgChan chan kafka.Message) {
	log.Info.Println("Subscribing to topic", r.configuration.Topic)
	for {
		msg, err := r.ReadMessage(ctx)

		if err != nil {
			log.Error.Println("failed to read message from topic:", r.configuration.Topic, err)
		}
		if msg.Topic == r.configuration.Topic {
			log.Info.Println("Read new message with key: ", string(msg.Key))
			msgChan <- msg
		}
	}
}
