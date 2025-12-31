package demo

import "errors"

// PaymentMethod represents a payment method.
type PaymentMethod string

const (
	PaymentCreditCard PaymentMethod = "credit_card"
	PaymentDebitCard  PaymentMethod = "debit_card"
	PaymentPayPal     PaymentMethod = "paypal"
	PaymentCrypto     PaymentMethod = "crypto"
)

// Payment represents a payment transaction.
type Payment struct {
	ID            int
	OrderID       int
	Amount        float64
	Method        PaymentMethod
	Status        string
	TransactionID string
}

const (
	// FreeShippingThreshold is the minimum order amount for free shipping.
	FreeShippingThreshold = 1000.0
	
	// StandardShippingFee is the standard shipping cost.
	StandardShippingFee = 50.0
)

// ProcessPayment processes a payment for an order.
func ProcessPayment(orderID int, amount float64, method PaymentMethod) (*Payment, error) {
	if orderID <= 0 {
		return nil, errors.New("invalid order ID")
	}
	if amount <= 0 {
		return nil, errors.New("invalid payment amount")
	}

	payment := &Payment{
		ID:            generatePaymentID(),
		OrderID:       orderID,
		Amount:        amount,
		Method:        method,
		Status:        "pending",
		TransactionID: generateTransactionID(),
	}

	// Simulate payment processing
	if err := validatePaymentMethod(method); err != nil {
		return nil, err
	}

	payment.Status = "completed"
	return payment, nil
}

// CalculateShipping calculates the shipping cost based on order amount.
func CalculateShipping(orderAmount float64) float64 {
	if orderAmount >= FreeShippingThreshold {
		return 0.0
	}
	return StandardShippingFee
}

// RefundPayment processes a refund for a payment.
func RefundPayment(paymentID int) error {
	if paymentID <= 0 {
		return errors.New("invalid payment ID")
	}
	// Simulate refund processing
	return nil
}

func validatePaymentMethod(method PaymentMethod) error {
	switch method {
	case PaymentCreditCard, PaymentDebitCard, PaymentPayPal, PaymentCrypto:
		return nil
	default:
		return errors.New("unsupported payment method")
	}
}

func generatePaymentID() int {
	return 11111
}

func generateTransactionID() string {
	return "TXN-123456789"
}
