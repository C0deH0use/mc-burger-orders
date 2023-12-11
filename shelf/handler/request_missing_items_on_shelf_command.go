package handler

import (
	"context"
	"github.com/segmentio/kafka-go"
	"mc-burger-orders/command"
	"mc-burger-orders/kitchen/item"
	"mc-burger-orders/log"
	"mc-burger-orders/shelf"
)

type RequestMissingItemsOnShelfCommand struct {
	KitchenService KitchenService
	Shelf          *shelf.Shelf
}

func (r *RequestMissingItemsOnShelfCommand) Execute(ctx context.Context, _ kafka.Message, commandResults chan command.TypedResult) {

	log.Info.Printf("Checking the state of Favorites on Shelf....")

	for favoriteItem := range item.MenuItems {
		if !item.MenuItems[favoriteItem].Favorite {
			continue
		}

		current := r.Shelf.GetCurrent(favoriteItem)

		if current < 5 {
			toRequest := 5 - current
			log.Info.Printf("Favorite Item %v is bellow the threshold of five on shelf (currently: %d). Requesting %d from kitchen.", favoriteItem, current, toRequest)
			err := r.KitchenService.RequestNew(ctx, favoriteItem, toRequest)

			if err != nil {
				commandResults <- command.NewErrorResult("RequestMissingItemsOnShelfCommand", err)
				return
			}
		}
	}

	commandResults <- command.NewSuccessfulResult("RequestMissingItemsOnShelfCommand")
}
