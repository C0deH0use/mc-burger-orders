package command

import (
	"fmt"
	om "mc-burger-orders/order/model"
)

type Command interface {
	Execute() (*om.Order, error)
}

type CommandHandler interface {
	Execute(c Command) (*om.Order, error)
}

type DefaultHandler struct {
}

func (receiver *DefaultHandler) Execute(c Command) (*om.Order, error) {
	fmt.Println("About to execute following command", c)

	result, err := c.Execute()
	if err != nil {
		fmt.Println("While executing command", c, "following error occurred", err.Error())
		return nil, err
	}

	fmt.Println("Command finished successfully")
	return result, nil

}
