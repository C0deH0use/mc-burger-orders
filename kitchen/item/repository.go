package item

import "fmt"

func IsKnownItem(item string) error {
	if _, exists := MenuItems[item]; !exists {
		err := fmt.Errorf("unknown item %qx", item)
		return err
	}

	return nil
}

func IsItemReady(item string) (bool, error) {
	if i, exists := MenuItems[item]; exists {
		return i.InstantReady, nil
	}

	err := fmt.Errorf("unknown item %q, cannot determin its readines status", item)
	return false, err
}
