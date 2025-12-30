package payment

// CalculateShipping calculates shipping cost.
// Intentionally uses 1000 to test inconsistency detection.
func CalculateShipping(amount float64) float64 {
	if amount >= 1000 { // Doc says 500, this is 1000
		return 0
	}
	return 10
}

// CalculateDiscount calculates discount for a given price.
func CalculateDiscount(price float64, isVIP bool) float64 {
	if isVIP {
		return price * 0.8 // 20% discount
	}
	return price
}
