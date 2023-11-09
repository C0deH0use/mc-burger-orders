package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/segmentio/kafka-go"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type KitchenRequestService interface {
	Request(itemName string, quantity int)
	RequestForOrder(ctx context.Context, itemName string, quantity int, orderNumber int64) error
}

type KitchenService struct {
	conn   *kafka.Conn
	writer *kafka.Writer
}

func NewKitchenServiceEnv() *KitchenService {
	kafkaAddressEnvVal := os.Getenv("KAFKA_ADDRESS")
	kafkaAddress := strings.Split(kafkaAddressEnvVal, "")
	topic := os.Getenv("KAFKA_TOPICS__KITCHEN_REQUESTS_TOPIC_NAME")

	return NewKitchenServiceFrom(kafkaAddress, topic)
}

func NewKitchenServiceFrom(kafkaAddress []string, topic string) *KitchenService {
	if len(kafkaAddress) == 0 {
		log.Fatal("missing at least one Kafka Address")
	}

	conn, err := kafka.Dial("tcp", kafkaAddress[0])
	if err != nil {
		log.Fatal("failed to dial leader:", err)
	}

	createTopic(conn, topic)

	writer := kafka.Writer{
		Addr:                   kafka.TCP(kafkaAddress...),
		Topic:                  topic,
		Balancer:               &kafka.LeastBytes{},
		RequiredAcks:           kafka.RequireOne,
		AllowAutoTopicCreation: true,
		ReadTimeout:            5 * time.Second,
		WriteTimeout:           5 * time.Second,
		MaxAttempts:            5,
	}
	return &KitchenService{conn: conn, writer: &writer}
}

func createTopic(conn *kafka.Conn, topicName string) {
	topicConfig := kafka.TopicConfig{Topic: topicName, NumPartitions: 1, ReplicationFactor: 1}

	err := conn.CreateTopics(topicConfig)
	if err != nil {
		log.Fatal("failed to create new topic", topicConfig, ".Error reason:", err.Error())
	}
}

// Request DEPRECATED: request method is going to be removed in the upcoming change in favor for: RequestForOrder
func (s *KitchenService) Request(itemName string, quantity int) {
	msgValue, err := json.Marshal(NewKitchenRequestMessage(itemName, quantity))
	if err != nil {
		log.Fatal("failed to convert message details to bytes", err)
	}
	msg := kafka.Message{
		Value: msgValue,
	}
	err = s.writer.WriteMessages(context.Background(), msg)
	if err != nil {
		log.Fatal("failed to set write deadline limit:", err)
	}

	defer func() {
		if err := s.writer.Close(); err != nil {
			log.Fatal("failed to close writer:", err)
		}
	}()
}

func (s *KitchenService) RequestForOrder(ctx context.Context, itemName string, quantity int, orderNumber int64) error {
	headers := make([]kafka.Header, 0)
	headers = append(headers, kafka.Header{Key: "order", Value: []byte(strconv.FormatInt(orderNumber, 10))})
	message := NewKitchenRequestMessage(itemName, quantity)
	msgValue, err := json.Marshal(message)
	if err != nil {
		err = fmt.Errorf("failed to convert message details to bytes. Reason: %s", err)
		return err
	}

	fmt.Println("Sending message with value", string(msgValue))
	msg := kafka.Message{
		Headers: headers,
		Value:   msgValue,
	}
	s.sendMessage(ctx, msg)

	return nil
}

func (s *KitchenService) sendMessage(rootCtx context.Context, messages ...kafka.Message) {
	var err error
	const retries = 3
	log.Println("Attempt to send", len(messages), "message(s)")
	for i := 0; i < retries; i++ {
		ctx, cancel := context.WithTimeout(rootCtx, 10*time.Second)
		defer cancel()

		// attempt to create topic prior to publishing the message
		err = s.writer.WriteMessages(ctx, messages...)
		if errors.Is(err, kafka.LeaderNotAvailable) || errors.Is(err, context.DeadlineExceeded) {
			log.Println("Message(s) where not send successfully", err, ". Waiting 250 milliseconds and will attempt for the", i+1, "time")

			time.Sleep(time.Millisecond * 250)
			continue
		}

		if err != nil {
			log.Fatal("unexpected error", err)
		}

		log.Println("Message(s) send successfully on attempt", i)
		break
	}

	if err := s.writer.Close(); err != nil {
		log.Fatal("failed to close writer:", err)
	}
}
