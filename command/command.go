package command

import (
	"context"
	"github.com/segmentio/kafka-go"
)

type Command interface {
	Execute(ctx context.Context, message kafka.Message, result chan TypedResult)
}
