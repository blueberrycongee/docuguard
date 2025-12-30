package payment

// CalculateShipping 计算运费
// 这里故意写成 1000 来测试不一致检测
func CalculateShipping(amount float64) float64 {
	if amount >= 1000 { // 文档说 500，这里是 1000
		return 0
	}
	return 10
}

// CalculateDiscount 计算折扣
func CalculateDiscount(price float64, isVIP bool) float64 {
	if isVIP {
		return price * 0.8 // 8 折
	}
	return price
}
