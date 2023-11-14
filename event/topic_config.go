package event

import (
	"fmt"
	"github.com/segmentio/kafka-go"
	"mc-burger-orders/log"
)

type TopicConfigs struct {
	Controller        string
	Brokers           []string
	Topic             string
	NumPartitions     int
	ReplicationFactor int
	AutoCreateTopic   bool
}

func (c *TopicConfigs) ConnectToBroker() *kafka.Conn {
	log.Warning.Println("Selecting one of the brokers in configuration...")
	brokerAddress := c.Brokers[0]
	conn, err := kafka.Dial("tcp", brokerAddress)
	if err != nil {
		log.Error.Fatal("failed to dial leader:", brokerAddress, err)
	}
	log.Warning.Println("Successfully dialed into broker", brokerAddress)
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
