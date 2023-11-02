package model

import (
	"github.com/stretchr/testify/assert"
	"mc-burger-orders/item"
	"testing"
)

func TestOrder_UpdateStatusWhenPackingItemsAndStillSomeAreMissingToCompleteOrder(t *testing.T) {
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
	count := order.getItemsCount(order.PackedItems)

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
	count := order.getItemsCount(order.PackedItems)

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
	count := order.getItemsCount(order.PackedItems)

	// then
	assert.Equal(t, 5, count)
}
