package command

import (
	"context"
	"mc-burger-orders/log"
)

type Command interface {
	Execute(ctx context.Context) (bool, error)
}

type ExecutionHandler interface {
	Execute(c Command) (bool, error)
}

type DefaultHandler struct {
	Events map[string][]Command
}

func (r *DefaultHandler) Register(event string, commands ...Command) {
	v, ok := r.Events[event]
	if ok {
		v = append(v, commands...)
	} else {
		r.Events[event] = commands
	}
}

func (r *DefaultHandler) Execute(c Command) (bool, error) {
	log.Info.Println("About to execute following command", c)

	result, err := c.Execute(context.Background())
	if err != nil {
		log.Error.Println("While executing command", c, "following error occurred", err.Error())
		return false, err
	}

	log.Info.Println("Command finished successfully")
	return result, nil

}
