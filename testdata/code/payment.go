package payment

// CalculateShipping 计算运费
// 修改：满 800 免运费
func CalculateShipping(amount float64) float64 {
	if amount >= 800 { // 修改为 800
		return 0
	}
	return 15 // 运费也调整了
}

// CalculateDiscount 计算折扣
func CalculateDiscount(price float64, isVIP bool) float64 {
	if isVIP {
		return price * 0.8 // 8 折
	}
	return price
}
