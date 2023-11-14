package dto

type StackUpdatedMessage []StackItemAddedMessage

type StackItemAddedMessage struct {
	ItemName string `json:"itemName"`
	Quantity int    `json:"quantity"`
}
