package dto

type StackItemAddedMessage struct {
	ItemName string `json:"itemName"`
	Quantity int    `json:"quantity"`
}
