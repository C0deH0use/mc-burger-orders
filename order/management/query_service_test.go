package management

import (
	"context"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"mc-burger-orders/kitchen/item"
	"mc-burger-orders/order"
	"mc-burger-orders/testing/utils"
	"testing"
	"time"
)

func TestIntegrationTest(t *testing.T) {
	utils.IntegrationTest(t)
	ctx := context.Background()

	mongoContainer, database = utils.TestWithMongo(t, ctx)

	collectionDb = database.Collection("orders")

	t.Run("should return Requested or InProgress orders when records are missing packed items", shouldReturnOrdersFulfillingQueryCriteriaWhenSuchRecordsExists)
	t.Run("should skip orders when records were last updated before the threshold", shouldSkipOrdersWhenRecordsWereLastUpdatedBeforeTheThreshold)
	t.Run("should skip orders when records are not missing any items but in the questioning status", shouldSkipOrdersWhenRecordsAreNotMissingAnyItemsButInTheQuestioningStatus)

	t.Cleanup(func() {
		t.Log("Running Clean UP code")
		utils.TerminateMongo(t, ctx, mongoContainer)
	})
}

func shouldReturnOrdersFulfillingQueryCriteriaWhenSuchRecordsExists(t *testing.T) {
	// given
	currentTime := time.Now()

	expectedOrders := []interface{}{
		order.Order{OrderNumber: 999, CustomerId: 10, Items: []item.Item{{Name: "hamburger", Quantity: 1}, {Name: "fries", Quantity: 1}}, Status: order.Ready, ModifiedAt: currentTime},
		order.Order{OrderNumber: 1000, CustomerId: 1, Items: []item.Item{{Name: "hamburger", Quantity: 1}, {Name: "fries", Quantity: 1}}, Status: order.Requested, ModifiedAt: currentTime.Add(time.Minute * -3)},
		order.Order{OrderNumber: 1002, CustomerId: 3, Items: []item.Item{{Name: "cheeseburger", Quantity: 2}, {Name: "hamburger", Quantity: 3}}, PackedItems: []item.Item{{Name: "hamburger", Quantity: 2}}, Status: order.InProgress, ModifiedAt: currentTime.Add(time.Minute * -10)},
	}
	utils.DeleteMany(t, collectionDb, bson.D{})
	utils.InsertMany(t, collectionDb, expectedOrders)

	sut := NewOrderQueryService(NewOrderRepository(database))

	// when
	results, err := sut.FetchOrdersForPacking(context.Background())

	// then
	assert.Nil(t, err)
	assert.Len(t, results, 2)

	// and
	orderIds := make([]int64, 0)
	for _, orderResult := range results {
		orderIds = append(orderIds, orderResult.OrderNumber)
	}

	assert.Len(t, orderIds, 2)
	assert.Contains(t, orderIds, int64(1000))
	assert.Contains(t, orderIds, int64(1002))
}

func shouldSkipOrdersWhenRecordsAreNotMissingAnyItemsButInTheQuestioningStatus(t *testing.T) {
	// given
	currentTime := time.Now()

	expectedOrders := []interface{}{
		order.Order{OrderNumber: 999, CustomerId: 10, Items: []item.Item{{Name: "hamburger", Quantity: 1}, {Name: "fries", Quantity: 1}}, Status: order.Ready, ModifiedAt: currentTime},
		order.Order{OrderNumber: 1000, CustomerId: 1, Items: []item.Item{{Name: "hamburger", Quantity: 1}}, PackedItems: []item.Item{{Name: "hamburger", Quantity: 1}}, Status: order.Requested, ModifiedAt: currentTime.Add(time.Minute * -3)},
		order.Order{OrderNumber: 1002, CustomerId: 3, Items: []item.Item{{Name: "cheeseburger", Quantity: 2}, {Name: "hamburger", Quantity: 3}}, PackedItems: []item.Item{{Name: "cheeseburger", Quantity: 2}, {Name: "hamburger", Quantity: 3}}, Status: order.InProgress, ModifiedAt: currentTime.Add(time.Minute * -10)},
	}
	utils.DeleteMany(t, collectionDb, bson.D{})
	utils.InsertMany(t, collectionDb, expectedOrders)

	sut := NewOrderQueryService(NewOrderRepository(database))

	// when
	results, err := sut.FetchOrdersForPacking(context.Background())

	// then
	assert.Nil(t, err)
	assert.Len(t, results, 0)
}

func shouldSkipOrdersWhenRecordsWereLastUpdatedBeforeTheThreshold(t *testing.T) {
	// given
	currentTime := time.Now()

	expectedOrders := []interface{}{
		order.Order{OrderNumber: 999, CustomerId: 10, Items: []item.Item{{Name: "hamburger", Quantity: 1}, {Name: "fries", Quantity: 1}}, Status: order.Ready, ModifiedAt: currentTime},
		order.Order{OrderNumber: 1000, CustomerId: 1, Items: []item.Item{{Name: "hamburger", Quantity: 1}, {Name: "fries", Quantity: 1}}, Status: order.Requested, ModifiedAt: currentTime.Add(time.Minute * -1)},
		order.Order{OrderNumber: 1002, CustomerId: 3, Items: []item.Item{{Name: "cheeseburger", Quantity: 2}, {Name: "hamburger", Quantity: 3}}, PackedItems: []item.Item{{Name: "hamburger", Quantity: 2}}, Status: order.InProgress, ModifiedAt: currentTime.Add(time.Second * -115)},
	}
	utils.DeleteMany(t, collectionDb, bson.D{})
	utils.InsertMany(t, collectionDb, expectedOrders)

	sut := NewOrderQueryService(NewOrderRepository(database))

	// when
	results, err := sut.FetchOrdersForPacking(context.Background())

	// then
	assert.Nil(t, err)
	assert.Len(t, results, 0)
}
