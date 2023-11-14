package stack

import (
	"fmt"
)

type Stack struct {
	kitchenStack map[string]int
}

func NewStack(ks map[string]int) *Stack {
	return &Stack{kitchenStack: ks}
}

func CleanStack() map[string]int {
	return map[string]int{
		"hamburger":       0,
		"cheeseburger":    0,
		"double-cheese":   0,
		"mc-chicken":      0,
		"mr-chicken-wrap": 0,
		"spicy-stripes":   0,
		"hot-wings":       0,
		"fries":           0,
	}
}

func (s *Stack) Add(i string) {
	s.kitchenStack[i] += 1
}
func (s *Stack) AddMany(i string, q int) {
	s.kitchenStack[i] += q
}

func (s *Stack) GetCurrent(i string) int {
	return s.kitchenStack[i]
}

func (s *Stack) Take(i string, q int) error {
	itemQuantity := s.kitchenStack[i]
	if itemQuantity < q {
		err := fmt.Errorf("not enought in Stack of type %q. Requered %d, but only %d available", i, itemQuantity, q)
		return err
	}

	s.kitchenStack[i] -= q
	return nil
}
