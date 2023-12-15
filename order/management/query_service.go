package management

import (
	"context"
	"mc-burger-orders/order"
)

type OrderQueryService interface {
	FetchOrdersForPacking(ctx context.Context) ([]order.Order, error)
}

type OrderQueryServiceImp struct {
	Repository PackingOrderItemsRepository
}

func (o OrderQueryServiceImp) FetchOrdersForPacking(ctx context.Context) ([]order.Order, error) {
	dbRecords, err := o.Repository.FetchOrdersWithMissingPackedItems(ctx)
	if err != nil {
		return nil, err
	}

	r := make([]order.Order, 0)
	for _, record := range dbRecords {
		if len(record.GetMissingItems()) > 0 {
			r = append(r, record)
		}
	}
	return r, nil
}

func NewOrderQueryService(r PackingOrderItemsRepository) OrderQueryService {
	return &OrderQueryServiceImp{r}
}
