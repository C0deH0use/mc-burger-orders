package kitchen

import (
	"mc-burger-orders/event"
	"mc-burger-orders/stack"
)

func TopicConfigsFromEnv() *event.TopicConfigs {
	return stack.TopicConfigsFromEnv()
}
