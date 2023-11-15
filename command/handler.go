package command

import (
	"fmt"
	"github.com/segmentio/kafka-go"
	"mc-burger-orders/log"
	"mc-burger-orders/utils"
)

type Handler interface {
	AddCommands(event string, commands ...Command)
	GetCommands(message kafka.Message) ([]Command, error)
	Handle(message kafka.Message) (bool, error)
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
	err = fmt.Errorf("failed to find command handler for messages of topic: %s", message.Topic)
	log.Error.Fatalln(err)
	return nil, err
}

func (o *DefaultCommandHandler) Handle(message kafka.Message) (bool, error) {
	commands, err := o.GetCommands(message)
	if err != nil {
		log.Error.Println(err)
		return false, err
	}

	result := false
	log.Info.Printf("Message will be executed on %d command(s)\n", len(commands))
	for _, command := range commands {

		result, err = o.Execute(command)
		if err != nil {
			log.Error.Println("While executing cmd", command, "following error occurred", err.Error())
			return false, err
		}
	}

	log.Info.Println("Command(s) finished successfully, with result -", result)
	return result, nil
}
