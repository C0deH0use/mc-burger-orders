package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/segmentio/kafka-go"
	"github.com/spf13/cast"
	"mc-burger-orders/log"
	"os"
	"strconv"
	"strings"
	"time"
)

type KitchenRequestService interface {
	RequestForOrder(ctx context.Context, itemName string, quantity int, orderNumber int64) error
}

type KitchenService struct {
	conn          *kafka.Conn
	configuration *KitchenServiceConfigs
}

type KitchenServiceConfigs struct {
	Controller        string
	Brokers           []string
	Topic             string
	NumPartitions     int
	ReplicationFactor int
}

func KitchenServiceConfigsFromEnv() KitchenServiceConfigs {
	kafkaAddressEnvVal := os.Getenv("KAFKA_ADDRESS")
	kafkaAddress := strings.Split(kafkaAddressEnvVal, ",")
	topic := os.Getenv("KAFKA_TOPICS__KITCHEN_REQUESTS_TOPIC_NAME")

	numPartitions := 1
	numPartitionsVal := os.Getenv("KAFKA_TOPICS__KITCHEN_REQUESTS_TOPIC_NUMBER_OF_PARTITIONS")
	replicationFactor := 1
	replicationFactorVal := os.Getenv("KAFKA_TOPICS__KITCHEN_REQUESTS_TOPIC_REPLICA_FACTOR")

	if len(numPartitionsVal) > 0 {
		numPartitions = cast.ToInt(numPartitionsVal)
	}
	if len(replicationFactorVal) > 0 {
		replicationFactor = cast.ToInt(replicationFactorVal)
	}

	return KitchenServiceConfigs{
		Brokers:           kafkaAddress,
		Topic:             topic,
		NumPartitions:     numPartitions,
		ReplicationFactor: replicationFactor,
	}
}

func NewKitchenServiceFrom(config KitchenServiceConfigs) *KitchenService {
	if len(config.Brokers) == 0 {
		log.Error.Panicln("missing at least one Kafka Address")
	}

	conn := config.connectToBroker()
	createTopic(conn, config)
	return &KitchenService{conn: conn, configuration: &config}
}

func (c *KitchenServiceConfigs) connectToBroker() *kafka.Conn {
	log.Info.Println("Selecting one of the brokers in configuration...")
	brokerAddress := c.Brokers[0]
	conn, err := kafka.Dial("tcp", brokerAddress)
	if err != nil {
		log.Error.Fatal("failed to dial leader:", brokerAddress, err)
	}
	log.Info.Println("Successfully dialed into broker", brokerAddress)
	controller, err := conn.Controller()
	controllerAddress := fmt.Sprintf("%s:%d", controller.Host, controller.Port)

	if brokerAddress != controllerAddress {
		log.Warning.Println("The configured broker is not the controller configured in the cluster... Switching to", controllerAddress)
		brokerAddress = controllerAddress
		conn, err = kafka.Dial("tcp", brokerAddress)
		if err != nil {
			log.Error.Panicln("failed to dial leader:", brokerAddress, err)
		}
	}
	c.Controller = brokerAddress
	return conn
}

func createTopic(conn *kafka.Conn, configs KitchenServiceConfigs) {
	topicConfig := kafka.TopicConfig{Topic: configs.Topic, NumPartitions: configs.NumPartitions, ReplicationFactor: configs.ReplicationFactor}

	err := conn.CreateTopics(topicConfig)
	if err != nil {
		log.Error.Panicln("failed to create new Topic", topicConfig, ".Error reason:", err.Error())
	}
}

func (s *KitchenService) createWriter() *kafka.Writer {
	return &kafka.Writer{
		Addr:                   kafka.TCP(s.configuration.Controller),
		Topic:                  s.configuration.Topic,
		Balancer:               &kafka.LeastBytes{},
		RequiredAcks:           kafka.RequireOne,
		AllowAutoTopicCreation: true,
		ReadTimeout:            5 * time.Second,
		WriteTimeout:           5 * time.Second,
		MaxAttempts:            5,
	}
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

	log.Info.Println("Sending message with value", string(msgValue))
	msg := kafka.Message{
		Headers: headers,
		Value:   msgValue,
	}
	err = s.sendMessage(ctx, msg)
	if err != nil {
		return err
	}
	return nil
}

func (s *KitchenService) sendMessage(rootCtx context.Context, messages ...kafka.Message) error {
	writer := s.createWriter()
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
