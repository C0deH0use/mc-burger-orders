package command

import (
	"context"
	"mc-burger-orders/log"
	"reflect"
)

type Dispatcher interface {
	Execute(c Command) (bool, error)
}

type DefaultDispatcher struct{}

func (r *DefaultDispatcher) Execute(c Command) (bool, error) {
	log.Info.Println("About to execute following command", reflect.TypeOf(c))

	result, err := c.Execute(context.Background())
	if err != nil {
		log.Error.Printf("While executing command %v following error occurred %v\n", c, err.Error())
		return false, err
	}

	log.Info.Println("Command finished successfully")
	return result, nil
}
