package order

import (
	"context"
	"fmt"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"mc-burger-orders/command"
	"net/http"
	"testing"
)

func TestOrderCollectedCommand_Execute(t *testing.T) {
	t.Run("should emit status update when client collects ready order", shouldEmitStatusUpdateWhenClientCollectsReadyOrder)
	t.Run("should not emit status update when order is not yet ready", shouldNotEmitStatusUpdateWhenOrderIsNotYetReady)
	t.Run("should not emit status update when order by number does not exists", shouldNotEmitStatusUpdateWhenOrderByNumberDoesNotExists)
	t.Run("should not emit status update when order was already collected", shouldNotEmitStatusUpdateWhenOrderWasAlreadyCollected)
}

func shouldEmitStatusUpdateWhenClientCollectsReadyOrder(t *testing.T) {
	// given
	stubRepository := GivenRepository()
	stubService := NewStubService()

	stubRepository.ReturnFetchByOrderNumber(&Order{OrderNumber: expectedOrderNumber, Status: Ready})

	sut := &OrderCollectedCommand{OrderNumber: expectedOrderNumber, Repository: stubRepository, StatusEmitter: stubService}
	commandResults := make(chan command.TypedResult)

	// when
	go sut.Execute(context.Background(), kafka.Message{}, commandResults)

	// then
	result := <-commandResults
	assert.True(t, result.Result)

	upsertArgs := stubRepository.GetUpsertArgs()
	assert.Len(t, upsertArgs, 1)
	assert.Equal(t, Collected, upsertArgs[0].Status)

	// and
	assert.True(t, stubService.HaveBeenCalledWith(StatusUpdateMatchingFnc(Collected)))
}

func shouldNotEmitStatusUpdateWhenOrderIsNotYetReady(t *testing.T) {
	// given
	stubRepository := GivenRepository()
	stubService := NewStubService()

	stubRepository.ReturnFetchByOrderNumber(&Order{OrderNumber: expectedOrderNumber, Status: InProgress})

	sut := &OrderCollectedCommand{OrderNumber: expectedOrderNumber, Repository: stubRepository, StatusEmitter: stubService}
	commandResults := make(chan command.TypedResult)

	// when
	go sut.Execute(context.Background(), kafka.Message{}, commandResults)

	// then
	result := <-commandResults
	assert.False(t, result.Result)
	assert.Equal(t, "requested order is yet ready for collection", result.Error.ErrorMessage)
	assert.Equal(t, http.StatusPreconditionRequired, result.Error.HttpResponse)

	assert.Len(t, stubRepository.GetUpsertArgs(), 0)
	assert.Equal(t, 0, stubService.CalledCnt())
}

func shouldNotEmitStatusUpdateWhenOrderByNumberDoesNotExists(t *testing.T) {
	// given
	stubRepository := GivenRepository()
	stubService := NewStubService()

	stubRepository.ReturnError(fmt.Errorf("error fetching order"))

	sut := &OrderCollectedCommand{OrderNumber: expectedOrderNumber, Repository: stubRepository, StatusEmitter: stubService}
	commandResults := make(chan command.TypedResult)

	// when
	go sut.Execute(context.Background(), kafka.Message{}, commandResults)

	// then
	result := <-commandResults
	assert.False(t, result.Result)
	assert.Equal(t, "failed to find order by order number. Reason: error fetching order", result.Error.ErrorMessage)
	assert.Equal(t, http.StatusNotFound, result.Error.HttpResponse)

	assert.Len(t, stubRepository.GetUpsertArgs(), 0)
	assert.Equal(t, 0, stubService.CalledCnt())
}

func shouldNotEmitStatusUpdateWhenOrderWasAlreadyCollected(t *testing.T) {
	// given
	stubRepository := GivenRepository()
	stubService := NewStubService()

	stubRepository.ReturnFetchByOrderNumber(&Order{OrderNumber: expectedOrderNumber, Status: Collected})

	sut := &OrderCollectedCommand{OrderNumber: expectedOrderNumber, Repository: stubRepository, StatusEmitter: stubService}
	commandResults := make(chan command.TypedResult)

	// when
	go sut.Execute(context.Background(), kafka.Message{}, commandResults)

	// then
	result := <-commandResults
	assert.False(t, result.Result)
	assert.Equal(t, "requested order already is collected", result.Error.ErrorMessage)
	assert.Equal(t, http.StatusPreconditionFailed, result.Error.HttpResponse)

	assert.Len(t, stubRepository.GetUpsertArgs(), 0)
	assert.Equal(t, 0, stubService.CalledCnt())
}
