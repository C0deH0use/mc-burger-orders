package command

import (
	"context"
)

type Command interface {
	Execute(ctx context.Context) (bool, error)
}
