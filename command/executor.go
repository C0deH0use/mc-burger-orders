package command

import (
	"context"
	"github.com/segmentio/kafka-go"
	"mc-burger-orders/log"
	"reflect"
	"time"
)

type Dispatcher interface {
	Execute(c Command, message kafka.Message, result chan TypedResult)
}

type DefaultDispatcher struct{}

func (r *DefaultDispatcher) Execute(c Command, message kafka.Message, result chan TypedResult) {
	log.Info.Println("About to execute following command", reflect.TypeOf(c))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	c.Execute(ctx, message, result)
}
