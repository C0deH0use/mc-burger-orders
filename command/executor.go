package command

import (
	"context"
	"mc-burger-orders/log"
)

type Dispatcher interface {
	Execute(c Command) (bool, error)
}

type DefaultDispatcher struct{}

func (r *DefaultDispatcher) Execute(c Command) (bool, error) {
	log.Info.Println("About to execute following command", c)

	result, err := c.Execute(context.Background())
	if err != nil {
		log.Error.Println("While executing command", c, "following error occurred", err.Error())
		return false, err
	}

	log.Info.Println("Command finished successfully")
	return result, nil
}
