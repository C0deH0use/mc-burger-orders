package event

import (
	"context"
	"errors"
	"github.com/segmentio/kafka-go"
	"mc-burger-orders/log"
	"time"
)

type DefaultWriter struct {
	conn          *kafka.Conn
	configuration *TopicConfigs
}

func NewTopicWriter(configuration *TopicConfigs) *DefaultWriter {
	if len(configuration.Brokers) == 0 {
		log.Error.Panicln("missing at least one Kafka Address")
	}

	conn := configuration.ConnectToBroker()
	configuration.CreateTopic(conn)
	return &DefaultWriter{conn: conn, configuration: configuration}
}

func (d *DefaultWriter) initializeWriter() *kafka.Writer {
	return &kafka.Writer{
		Addr:                   kafka.TCP(d.configuration.Controller),
		Topic:                  d.configuration.Topic,
		Balancer:               &kafka.LeastBytes{},
		RequiredAcks:           kafka.RequireOne,
		AllowAutoTopicCreation: true,
		ReadTimeout:            5 * time.Second,
		WriteTimeout:           5 * time.Second,
		MaxAttempts:            5,
	}
}

func (d *DefaultWriter) SendMessage(rootCtx context.Context, messages ...kafka.Message) error {
	writer := d.initializeWriter()
	var err error
	const retries = 3
	log.Info.Println("Attempt to send", len(messages), "message(s)")
	for i := 0; i < retries; i++ {
		ctx, cancel := context.WithTimeout(rootCtx, 10*time.Second)
		defer cancel()

		// attempt to create Topic prior to publishing the message
		err = writer.WriteMessages(ctx, messages...)
		if errors.Is(err, kafka.LeaderNotAvailable) || errors.Is(err, context.DeadlineExceeded) {
			log.Error.Println("Message(s) where not send successfully", err, ". Waiting 250 milliseconds and will attempt for the", i+1, "time")

			time.Sleep(time.Millisecond * 250)
			continue
		}

		if err != nil {
			log.Error.Println("unexpected error", err)
			return err
		}

		log.Info.Println("Message(s) send successfully on attempt", i)
		break
	}

	if err := writer.Close(); err != nil {
		log.Error.Println("failed to close writer:", err)
		return err
	}
	return nil
}