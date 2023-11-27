package model

import (
	"github.com/stretchr/testify/assert"
	"mc-burger-orders/kitchen/item"
	"testing"
)

func TestOrder_UpdateStatusWhenPackingItemsAndStillSomeAreMissingToCompleteOrder(t *testing.T) {
	// given
	order := Order{
		Items: []item.Item{
			{
				Name:     "some-item",
				Quantity: 2,
			},
			{
				Name:     "ice-cream",
				Quantity: 1,
			},
		},
		PackedItems: []item.Item{},
	}

	// when
	order.PackItem("ice-cream", 1)

	// then
	assert.Equal(t, OrderStatus("IN_PROGRESS"), order.Status)
}

func TestOrder_UpdateStatusWhenAllPackingItemsAndPresent(t *testing.T) {
	// given
	order := Order{
		Items: []item.Item{
			{
				Name:     "some-item",
				Quantity: 2,
			},
			{
				Name:     "ice-cream",
				Quantity: 1,
			},
		},
		PackedItems: []item.Item{
			{
				Name:     "some-item",
				Quantity: 1,
			},
			{
				Name:     "ice-cream",
				Quantity: 1,
			},
		},
	}

	// when
	order.PackItem("some-item", 1)

	// then
	assert.Equal(t, OrderStatus("READY"), order.Status)
}

func TestOrder_PackItemCountWhenZeroItemsPresent(t *testing.T) {
	// given
	order := Order{
		PackedItems: []item.Item{},
	}

	// when
	count := order.GetItemsCount(order.PackedItems)

	// then
	assert.Equal(t, 0, count)
}

func TestOrder_PackItemCountWhenOneItemWithQuantityPresent(t *testing.T) {
	// given
	order := Order{
		PackedItems: []item.Item{
			{
				Name:     "some-item",
				Quantity: 2,
			},
		},
	}

	// when
	count := order.GetItemsCount(order.PackedItems)

	// then
	assert.Equal(t, 2, count)
}

func TestOrder_PackItemCountWhenTwoItemsWithDifferentQuantityPresent(t *testing.T) {
	// given
	order := Order{
		PackedItems: []item.Item{
			{
				Name:     "some-item",
				Quantity: 2,
			},
			{
				Name:     "some-item2",
				Quantity: 3,
			},
		},
	}

	// when
	count := order.GetItemsCount(order.PackedItems)

	// then
	assert.Equal(t, 5, count)
}

func TestOrder_GetMissingItemsCount_WhenItemIsNotFullyReady(t *testing.T) {
	// given
	order := Order{
		Items: []item.Item{
			{
				Name:     "some-item",
				Quantity: 2,
			},
			{
				Name:     "ice-cream",
				Quantity: 1,
			},
		},
		PackedItems: []item.Item{
			{
				Name:     "some-item",
				Quantity: 2,
			},
		},
	}

	// when
	count, err := order.GetMissingItemsCount("ice-cream")

	// then
	assert.Nil(t, err)

	assert.Equal(t, 1, count)
}

func TestOrder_GetMissingItemsCount_WhenItemReady(t *testing.T) {
	// given
	order := Order{
		Items: []item.Item{
			{
				Name:     "some-item",
				Quantity: 2,
			},
			{
				Name:     "ice-cream",
				Quantity: 1,
			},
		},
		PackedItems: []item.Item{
			{
				Name:     "some-item",
				Quantity: 2,
			},
		},
	}

	// when
	count, err := order.GetMissingItemsCount("some-item")

	// then
	assert.Nil(t, err)

	assert.Equal(t, 0, count)
}

func TestOrder_GetMissingItemsCount_Error_When_AskingAboutNonExistingItem(t *testing.T) {
	// given
	order := Order{
		Items: []item.Item{
			{
				Name:     "some-item",
				Quantity: 2,
			},
			{
				Name:     "ice-cream",
				Quantity: 1,
			},
		},
		PackedItems: []item.Item{
			{
				Name:     "some-item",
				Quantity: 2,
			},
		},
	}

	// when
	count, err := order.GetMissingItemsCount("some-item2")

	// then
	assert.Equal(t, "could not find item `some-item2` on the order list", err.Error())

	assert.Equal(t, 1, count)
}

func TestOrder_GetMissingItems_WhenOrderReady(t *testing.T) {
	// given
	order := Order{
		Items: []item.Item{
			{
				Name:     "some-item",
				Quantity: 2,
			},
			{
				Name:     "ice-cream",
				Quantity: 1,
			},
		},
		PackedItems: []item.Item{
			{
				Name:     "some-item",
				Quantity: 2,
			},
			{
				Name:     "ice-cream",
				Quantity: 1,
			},
		},
	}

	// when
	result := order.GetMissingItems()

	// then
	assert.Len(t, result, 2)

	assert.Contains(t, result, item.Item{Name: "some-item", Quantity: 0})
	assert.Contains(t, result, item.Item{Name: "ice-cream", Quantity: 0})
}

func TestOrder_GetMissingItems_WhenOrderInProgress(t *testing.T) {
	// given
	order := Order{
		Items: []item.Item{
			{
				Name:     "some-item",
				Quantity: 2,
			},
			{
				Name:     "ice-cream",
				Quantity: 1,
			},
		},
		PackedItems: []item.Item{
			{
				Name:     "some-item",
				Quantity: 2,
			},
		},
	}

	// when
	result := order.GetMissingItems()

	// then
	assert.Len(t, result, 2)

	assert.Contains(t, result, item.Item{Name: "some-item", Quantity: 0})
	assert.Contains(t, result, item.Item{Name: "ice-cream", Quantity: 1})
}

func TestOrder_GetMissingItems_WhenOrderHasZeroItems(t *testing.T) {
	// given
	order := Order{
		Items: make([]item.Item, 0),
		PackedItems: []item.Item{
			{
				Name:     "some-item",
				Quantity: 2,
			},
		},
	}

	// when
	result := order.GetMissingItems()

	// then
	assert.Len(t, result, 0)
}

func TestOrder_GetMissingItems_WhenOrderHasZeroPackedItems(t *testing.T) {
	// given
	order := Order{
		Items: []item.Item{
			{
				Name:     "some-item",
				Quantity: 2,
			},
		},
		PackedItems: make([]item.Item, 0),
	}

	// when
	result := order.GetMissingItems()

	// then
	assert.Len(t, result, 1)

	assert.Contains(t, result, item.Item{Name: "some-item", Quantity: 2})

}
