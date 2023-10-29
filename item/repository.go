package item

import "fmt"

func KnownItem(item string) error {
	_, exists := items[item]
	if !exists {
		err := fmt.Errorf("unknown item %q, cannot determin its readines status", item)
		return err
	}

	return nil
}

func IsItemReady(item string) (bool, error) {
	i, exists := items[item]
	if !exists {
		err := fmt.Errorf("unknown item %q, cannot determin its readines status", item)
		return false, err
	}

	return i, nil
}

var items = map[string]bool{
	"hamburger":       false,
	"cheeseburger":    false,
	"double-cheese":   false,
	"mc-chicken":      false,
	"mr-chicken-wrap": false,
	"spicy-stripes":   false,
	"hot-wings":       false,
	"fries":           false,
	"coke":            true,
	"ice-cream":       true,
	"fanta":           true,
}
