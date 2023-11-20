package stack

import (
	"fmt"
	"github.com/spf13/cast"
	"mc-burger-orders/log"
	"sync"
)

type Stack struct {
	kitchenStack *sync.Map
}

func NewStack(ks *sync.Map) *Stack {
	return &Stack{kitchenStack: ks}
}

func NewEmptyStack() *Stack {
	return &Stack{kitchenStack: CleanStack()}
}

func CleanStack() *sync.Map {
	syncMap := &sync.Map{}
	syncMap.Store("hamburger", 0)
	syncMap.Store("cheeseburger", 0)
	syncMap.Store("double-cheese", 0)
	syncMap.Store("mc-chicken", 0)
	syncMap.Store("mr-chicken-wrap", 0)
	syncMap.Store("spicy-stripes", 0)
	syncMap.Store("hot-wings", 0)
	syncMap.Store("fries", 0)
	return syncMap
}

func (s *Stack) Add(i string) {
	if value, ok := s.kitchenStack.Load(i); ok {
		newVal := cast.ToInt(value) + 1
		s.kitchenStack.Store(i, newVal)
		log.Warning.Printf("Kitchen Stack| %v => %d", i, newVal)
		return
	}
	s.kitchenStack.Store(i, 1)
	log.Warning.Printf("Kitchen Stack| %v => %d", i, 1)

}
func (s *Stack) AddMany(i string, q int) {
	if value, ok := s.kitchenStack.Load(i); ok {
		newVal := cast.ToInt(value) + q
		s.kitchenStack.Store(i, newVal)
		log.Warning.Printf("Kitchen Stack| %v => %d", i, newVal)
		return
	}
	s.kitchenStack.Store(i, q)
	log.Warning.Printf("Kitchen Stack| %v => %d", i, q)
}

func (s *Stack) GetCurrent(i string) int {
	if value, ok := s.kitchenStack.Load(i); ok {
		return cast.ToInt(value)
	}
	return 0
}

func (s *Stack) Take(i string, q int) error {
	if value, ok := s.kitchenStack.Load(i); ok {
		exists := cast.ToInt(value)
		if exists < q {
			err := fmt.Errorf("not enought in Stack of type %q. Requered %d, but only %d available", i, exists, q)
			return err
		}

		newVal := exists - q
		s.kitchenStack.Store(i, newVal)
		log.Warning.Printf("Kitchen Stack| %v => %d", i, newVal)
		return nil
	}
	err := fmt.Errorf("unknown item `%v` requested", i)
	return err
}
