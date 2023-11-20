package kitchen

type ItemRequest struct {
	ItemName string `json:"itemName"`
	Quantity int    `json:"quantity"`
}
