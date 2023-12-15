package event

import (
	"fmt"
	"github.com/segmentio/kafka-go"
	"github.com/spf13/cast"
	"mc-burger-orders/log"
	"os"
	"strings"
	"time"
)

type TopicConfigs struct {
	Controller            string
	Brokers               []string
	Topic                 string
	NumPartitions         int
	Partition             int
	ReplicationFactor     int
	WaitMaxTime           time.Duration
	AwaitBetweenReadsTime time.Duration
	AutoCreateTopic       bool
}

func NewTopicConfig(topic string, partition int, numPartitionsVal string, replicationFactorVal string) *TopicConfigs {
	kafkaAddressEnvVal := os.Getenv("KAFKA_ADDRESS")
	kafkaAddress := strings.Split(kafkaAddressEnvVal, ",")

	numPartitions := 3
	replicationFactor := 1

	if len(numPartitionsVal) > 0 {
		numPartitions = cast.ToInt(numPartitionsVal)
	}
	if len(replicationFactorVal) > 0 {
		replicationFactor = cast.ToInt(replicationFactorVal)
	}

	return &TopicConfigs{
		Brokers:               kafkaAddress,
		Topic:                 topic,
		NumPartitions:         numPartitions,
		Partition:             partition,
		ReplicationFactor:     replicationFactor,
		WaitMaxTime:           2 * time.Second,
		AwaitBetweenReadsTime: 500 * time.Millisecond,
	}
}

func (c *TopicConfigs) ConnectToBroker() *kafka.Conn {
	brokerAddress := c.Brokers[0]

	log.Warning.Printf("Connecting to broker `%v` from topic: %v", c.Brokers, c.Topic)
	conn, err := kafka.Dial("tcp", brokerAddress)
	if err != nil {
		log.Error.Fatal("failed to dial leader:", brokerAddress, err)
	}
	controller, err := conn.Controller()
	if err != nil {
		log.Error.Fatal("failed to connect to leading controller:", err)
	}
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

func (c *TopicConfigs) CreateTopic(conn *kafka.Conn) {
	if !c.AutoCreateTopic {
		return
	}

	topicConfig := kafka.TopicConfig{Topic: c.Topic, NumPartitions: c.NumPartitions, ReplicationFactor: c.ReplicationFactor}

	err := conn.CreateTopics(topicConfig)
	if err != nil {
		log.Error.Panicln("failed to create new Topic", topicConfig, ".Error reason:", err.Error())
	}
}
