package dto

type KitchenRequestMessage struct {
	ItemName string `json:"itemName"`
	Quantity int    `json:"quantity"`
}

func NewKitchenRequestMessage(name string, quantity int) *KitchenRequestMessage {
	return &KitchenRequestMessage{ItemName: name, Quantity: quantity}
}
