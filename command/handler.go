package command

import (
	"github.com/segmentio/kafka-go"
	"mc-burger-orders/log"
	"mc-burger-orders/utils"
)

type Handler interface {
	AddCommands(event string, commands ...Command)
	GetHandledEvents() []string
	GetCommands(message kafka.Message) ([]Command, error)
	Handle(message kafka.Message, typedResult chan TypedResult)
}

type DefaultCommandHandler struct {
	DefaultDispatcher
	eventHandlers map[string][]Command
}

func NewCommandHandler() *DefaultCommandHandler {
	return &DefaultCommandHandler{
		eventHandlers: make(map[string][]Command),
	}
}

func (o *DefaultCommandHandler) GetHandledEvents() []string {
	events := make([]string, 0)
	for event := range o.eventHandlers {
		events = append(events, event)
	}

	return events
}

func (o *DefaultCommandHandler) AddCommands(event string, commands ...Command) {
	if storedCommands, ok := o.eventHandlers[event]; ok {
		storedCommands = append(storedCommands, commands...)
		o.eventHandlers[event] = storedCommands
		return
	}
	o.eventHandlers[event] = commands
}

func (o *DefaultCommandHandler) GetCommands(message kafka.Message) ([]Command, error) {
	eventType, err := utils.GetEventType(message)
	if err != nil {
		log.Error.Println(err.Error())
		return nil, err
	}

	if commands, ok := o.eventHandlers[eventType]; ok {
		return commands, nil
	}
	log.Warning.Printf("failed to find command handler in `%v` for messages of topic: %v", "DefaultCommandHandler", message.Topic)
	return make([]Command, 0), err
}

func (o *DefaultCommandHandler) Handle(message kafka.Message, commandResults chan TypedResult) {
	commands, err := o.GetCommands(message)
	if err != nil {
		commandResults <- NewErrorResult("GetCommandForMessage", err)
		close(commandResults)
		return
	}

	o.HandleCommands(message, commandResults, commands...)
	//for commandResult := range commandResults {
	//	if commandResult.Error != nil {
	//		log.Error.Println("While executing command", commandResult.Type, "following error occurred", commandResult.Error.Error())
	//	} else {
	//		log.Info.Println("Command", commandResult.Type, "finished successfully, with result -", commandResult.Result)
	//	}
	//}
}

func (o *DefaultCommandHandler) HandleCommands(message kafka.Message, commandResults chan TypedResult, commands ...Command) {
	log.Info.Printf("Message will be executed on %d command(s)\n", len(commands))
	for _, command := range commands {
		go o.Execute(command, message, commandResults)
	}
}
