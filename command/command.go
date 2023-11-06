package command

import (
	"context"
	"fmt"
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
	fmt.Println("About to execute following command", c)

	// run command as coroutine
	//commandResults := make(chan []bool)
	//go func() {
	//	result, err := c.Execute(context.Background())
	//	if err != nil {
	//		fmt.Println("While executing command", c, "following error occurred", err.Error())
	//	}
	//	result >> commandResults
	//}()

	result, err := c.Execute(context.Background())
	if err != nil {
		fmt.Println("While executing command", c, "following error occurred", err.Error())
		return false, err
	}

	fmt.Println("Command finished successfully")
	return result, nil

}
