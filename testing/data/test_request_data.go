package data

func AppendHamburgerItem(messageValue []map[string]any, quantity int) []map[string]any {
	return append(messageValue, map[string]any{
		"itemName": "hamburger",
		"quantity": quantity,
	})
}
func AppendCheeseBurgerItem(messageValue []map[string]any, quantity int) []map[string]any {
	return append(messageValue, map[string]any{
		"itemName": "cheeseburger",
		"quantity": quantity,
	})
}
func AppendSpicyStripesItem(messageValue []map[string]any, quantity int) []map[string]any {
	return append(messageValue, map[string]any{
		"itemName": "spicy-stripes",
		"quantity": quantity,
	})
}
