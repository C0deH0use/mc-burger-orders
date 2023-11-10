package event

import (
	"context"
	"fmt"
	"github.com/segmentio/kafka-go"
	"mc-burger-orders/command"
	"mc-burger-orders/log"
)

type Handler interface {
	GetCommand(message kafka.Message) (command.Command, error)
	Handle(message kafka.Message) (bool, error)
}

type DefaultEventHandler struct {
}

func (o *DefaultEventHandler) GetCommand(message kafka.Message) (command.Command, error) {
	eventType := string(message.Key)
	err := fmt.Errorf("default event handler is not implemented for message type: %s", eventType)
	log.Error.Fatalln(err)
	return nil, err
}

func (o *DefaultEventHandler) Handle(message kafka.Message) (bool, error) {
	cmd, err := o.GetCommand(message)
	if err != nil {
		log.Error.Println(err)
		return false, err
	}

	log.Info.Println("About to execute following cmd", cmd)

	result, err := cmd.Execute(context.Background())
	if err != nil {
		log.Error.Println("While executing cmd", cmd, "following error occurred", err.Error())
		return false, err
	}

	log.Info.Println("Command finished successfully")
	return result, nil
}
