package command

import (
	"context"
	"github.com/segmentio/kafka-go"
	"mc-burger-orders/log"
	"reflect"
)

type Dispatcher interface {
	Execute(c Command, message kafka.Message) (bool, error)
}

type DefaultDispatcher struct{}

func (r *DefaultDispatcher) Execute(c Command, message kafka.Message) (bool, error) {
	log.Info.Println("About to execute following command", reflect.TypeOf(c))

	result, err := c.Execute(context.Background(), message)
	if err != nil {
		log.Error.Printf("While executing command %v following error occurred %v\n", c, err.Error())
		return false, err
	}

	log.Info.Println("Command finished successfully")
	return result, nil
}
