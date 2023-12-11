package handler

import (
	"context"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	command2 "mc-burger-orders/command"
	"mc-burger-orders/shelf"
	"testing"
)

func TestRequestMissingItemsOnShelfCommand_Execute(t *testing.T) {

	t.Run("should request required amount when favorite item is fully missing on shelf", shouldRequestItemsWhenMissingOnShelf)
	t.Run("should request only the needed amount when favorite item is missing on shelf", shouldRequestItemsWhenBellowRequiredLimitOfItemsOnShelf)
	t.Run("should not request any when all favorite items are on shelf", shouldNotRequestAnyWhenAllFavoriteItemsAreOnShelf)
}

func shouldRequestItemsWhenMissingOnShelf(t *testing.T) {
	// given
	s := shelf.NewEmptyShelf()
	s.AddMany("cheeseburger", 5)
	s.AddMany("spicy-stripes", 5)
	s.AddMany("hot-wings", 5)
	s.AddMany("fries", 5)

	kitchenStub := shelf.NewShelfStubService()

	sut := RequestMissingItemsOnShelfCommand{
		Shelf:          s,
		KitchenService: kitchenStub,
	}
	commandResults := make(chan command2.TypedResult)

	// when
	go sut.Execute(context.Background(), kafka.Message{}, commandResults)

	// then
	commandResult := <-commandResults
	assert.True(t, commandResult.Result)
	assert.Nil(t, commandResult.Error)

	// and
	assert.Equal(t, 1, kitchenStub.CalledCnt())
	assert.True(t, kitchenStub.HaveBeenCalledWith(shelf.RequestMatchingFnc("hamburger", 5)))
	close(commandResults)
}

func shouldRequestItemsWhenBellowRequiredLimitOfItemsOnShelf(t *testing.T) {
	// given
	s := shelf.NewEmptyShelf()
	s.AddMany("hamburger", 1)
	s.AddMany("cheeseburger", 5)
	s.AddMany("spicy-stripes", 5)
	s.AddMany("hot-wings", 5)
	s.AddMany("fries", 5)

	kitchenStub := shelf.NewShelfStubService()

	sut := RequestMissingItemsOnShelfCommand{
		Shelf:          s,
		KitchenService: kitchenStub,
	}

	commandResults := make(chan command2.TypedResult)

	// when
	go sut.Execute(context.Background(), kafka.Message{}, commandResults)

	// then
	commandResult := <-commandResults
	assert.True(t, commandResult.Result)
	assert.Nil(t, commandResult.Error)

	// and
	assert.Equal(t, 1, kitchenStub.CalledCnt())
	assert.True(t, kitchenStub.HaveBeenCalledWith(shelf.RequestMatchingFnc("hamburger", 4)))

	close(commandResults)
}

func shouldNotRequestAnyWhenAllFavoriteItemsAreOnShelf(t *testing.T) {
	// given
	s := shelf.NewEmptyShelf()
	s.AddMany("hamburger", 6)
	s.AddMany("cheeseburger", 5)
	s.AddMany("spicy-stripes", 8)
	s.AddMany("hot-wings", 10)
	s.AddMany("fries", 5)

	kitchenStub := shelf.NewShelfStubService()

	sut := RequestMissingItemsOnShelfCommand{
		Shelf:          s,
		KitchenService: kitchenStub,
	}
	commandResults := make(chan command2.TypedResult)

	// when
	go sut.Execute(context.Background(), kafka.Message{}, commandResults)

	// then
	commandResult := <-commandResults
	assert.True(t, commandResult.Result)
	assert.Nil(t, commandResult.Error)

	// and
	assert.Zero(t, kitchenStub.CalledCnt())
	close(commandResults)
}
