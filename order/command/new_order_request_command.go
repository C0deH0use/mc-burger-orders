package command

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	itemRepository "mc-burger-orders/item"
	"mc-burger-orders/order/model"
	"mc-burger-orders/stack"
	"mc-burger-orders/utils"
	"net/http"
)

type NewOrderRequestCommand struct {
	Repository model.OrderRepository
	Stack      *stack.Stack
}

func (command *NewOrderRequestCommand) CreateNewOrder(c *gin.Context) {
	newOrder := model.NewOrder{}
	err := c.ShouldBindJSON(&newOrder)
	if err != nil {
		errorMessage := fmt.Sprintf("Schema Error. %s", err)
		fmt.Println("New Order request Error: ", errorMessage)
		c.JSON(http.StatusBadRequest, utils.ErrorPayload(errorMessage))
		return
	}
	err = validate(newOrder)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, utils.ErrorPayload(err.Error()))
		return
	}

	order := command.Repository.CreateOrder(newOrder)
	for _, item := range newOrder.Items {
		isReady, err := itemRepository.IsItemReady(item.Name)
		if err != nil {
			fmt.Println(err)
			c.JSON(http.StatusBadRequest, utils.ErrorPayload(err.Error()))
			return
		}

		if isReady {
			order.PackItem(item, item.Quantity)
		} else {
			amountInStock := command.Stack.GetCurrent(item.Name)
			if amountInStock == 0 {
				// request to kitchen
				fmt.Println("Sending Request to kitchen for", item.Quantity, "new", item.Name)
			} else {
				var itemTaken int
				if amountInStock > item.Quantity {
					itemTaken = item.Quantity
				} else {
					itemTaken = item.Quantity - amountInStock
					remaining := item.Quantity - itemTaken
					// request to kitchen
					fmt.Println("Sending Request to kitchen for", remaining, "new", item.Name)
				}
				err = command.Stack.Take(item.Name, itemTaken)

				if err != nil {
					errMsg := fmt.Sprintf("error when collecting '%d' item(s) '%s' from stack", item.Quantity, item.Name)
					c.JSON(http.StatusInternalServerError, utils.ErrorPayload(errMsg))
					return
				}

				order.PackItem(item, itemTaken)
			}
		}
	}
	c.JSON(http.StatusCreated, order)
}

func validate(c model.NewOrder) error {
	var errs []error
	for _, item := range c.Items {
		err := itemRepository.KnownItem(item.Name)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}
