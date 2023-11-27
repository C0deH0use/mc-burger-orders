package order

import "time"

type OrderNumber struct {
	Number    int64     `bson:"number"`
	CreatedAt time.Time `bson:"createdAt"`
}

func NewOrderNumber(number int64) OrderNumber {
	return OrderNumber{Number: number, CreatedAt: time.Now()}
}
