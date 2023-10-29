package item

type Item struct {
	Name     string `json:"name" binding:"required"`
	Quantity int    `json:"quantity" binding:"gt=0"`
}
