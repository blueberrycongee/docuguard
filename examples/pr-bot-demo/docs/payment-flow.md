# Payment Processing Guide

## Overview

This guide explains how to process payments and handle shipping costs in the system.

## Payment Methods

The system supports the following payment methods:
- **Credit Card** - `PaymentCreditCard`
- **Debit Card** - `PaymentDebitCard`
- **PayPal** - `PaymentPayPal`
- **Cryptocurrency** - `PaymentCrypto`

## Processing Payments

### ProcessPayment

Processes a payment for an order.

**Signature:**
```go
func ProcessPayment(orderID int, amount float64, method PaymentMethod) (*Payment, error)
```

**Parameters:**
- `orderID` (int): The order ID to process payment for
- `amount` (float64): The payment amount
- `method` (PaymentMethod): The payment method to use

**Returns:**
- `*Payment`: The payment transaction object
- `error`: Error if payment processing fails

**Example:**
```go
payment, err := ProcessPayment(67890, 149.99, PaymentCreditCard)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Payment processed: %s (Status: %s)\n", payment.TransactionID, payment.Status)
```

## Shipping Costs

### Free Shipping Policy

Orders qualify for **free shipping** when the order total reaches **$500 or more**.

### CalculateShipping

Calculates the shipping cost based on the order amount.

**Signature:**
```go
func CalculateShipping(orderAmount float64) float64
```

**Parameters:**
- `orderAmount` (float64): The total order amount

**Returns:**
- `float64`: The shipping cost (0.0 if free shipping applies)

**Shipping Rates:**
- Orders **$500+**: Free shipping
- Orders **under $500**: $50.00 standard shipping

**Example:**
```go
// Order under threshold
shipping := CalculateShipping(299.99)
fmt.Printf("Shipping cost: $%.2f\n", shipping) // Output: $50.00

// Order qualifies for free shipping
shipping = CalculateShipping(599.99)
fmt.Printf("Shipping cost: $%.2f\n", shipping) // Output: $0.00
```

## Refunds

### RefundPayment

Processes a refund for a completed payment.

**Signature:**
```go
func RefundPayment(paymentID int) error
```

**Parameters:**
- `paymentID` (int): The payment ID to refund

**Returns:**
- `error`: Error if refund processing fails

**Example:**
```go
err := RefundPayment(11111)
if err != nil {
    log.Fatal(err)
}
fmt.Println("Refund processed successfully")
```

**Note:** Refunds are processed back to the original payment method and may take 3-5 business days to appear.

## Best Practices

1. Always validate the payment amount before processing
2. Handle payment errors gracefully and provide clear error messages
3. Log all payment transactions for audit purposes
4. Verify order status before processing refunds
5. Consider the free shipping threshold when displaying total costs to users
