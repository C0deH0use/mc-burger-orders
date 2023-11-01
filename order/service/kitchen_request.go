package service

type KitchenRequestService interface {
	Request(itemName string, quantity int)
}

type KitchenService struct {
}

func (s *KitchenService) Request(itemName string, quantity int) {

}
