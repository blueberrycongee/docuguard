# Order Management Guide

## Overview

This guide explains how to manage orders in the system.

## Order Lifecycle

Orders go through the following states:
1. **Pending** - Order created, awaiting confirmation
2. **Confirmed** - Order confirmed, ready for processing
3. **Shipped** - Order has been shipped
4. **Delivered** - Order delivered to customer
5. **Cancelled** - Order cancelled by user or system

## Creating Orders

### CreateOrder

Creates a new order for a user with specified items.

**Signature:**
```go
func CreateOrder(userID int, items []OrderItem) (*Order, error)
```

**Parameters:**
- `userID` (int): The ID of the user placing the order
- `items` ([]OrderItem): List of items to include in the order

**Returns:**
- `*Order`: The created order object
- `error`: Error if validation fails

**Example:**
```go
items := []OrderItem{
    {ProductID: 1, Quantity: 2, Price: 29.99},
    {ProductID: 2, Quantity: 1, Price: 49.99},
}

order, err := CreateOrder(12345, items)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Order created: #%d, Total: $%.2f\n", order.ID, order.TotalPrice)
```

## Retrieving Orders

### GetOrder

Retrieves an order by its ID.

**Signature:**
```go
func GetOrder(orderID int) *Order
```

**Parameters:**
- `orderID` (int): The order's unique identifier

**Returns:**
- `*Order`: The order object

**Example:**
```go
order := GetOrder(67890)
fmt.Printf("Order Status: %s\n", order.Status)
```

**Note:** This function always returns an order object. Check the order status to verify it exists.

## Updating Orders

### UpdateOrderStatus

Updates the status of an existing order.

**Signature:**
```go
func UpdateOrderStatus(orderID int, status OrderStatus) error
```

**Parameters:**
- `orderID` (int): The order's unique identifier
- `status` (OrderStatus): The new status

**Valid Status Values:**
- `StatusPending`
- `StatusConfirmed`
- `StatusShipped`
- `StatusDelivered`
- `StatusCancelled`

**Example:**
```go
err := UpdateOrderStatus(67890, StatusShipped)
if err != nil {
    log.Fatal(err)
}
```

## Cancelling Orders

### CancelOrder

Cancels an order if it hasn't been shipped yet.

**Signature:**
```go
func CancelOrder(orderID int) error
```

**Parameters:**
- `orderID` (int): The order's unique identifier

**Returns:**
- `error`: Error if cancellation fails

**Example:**
```go
err := CancelOrder(67890)
if err != nil {
    log.Fatal(err)
}
```

**Important:** Orders can only be cancelled if they are in `Pending` or `Confirmed` status. Shipped or delivered orders cannot be cancelled.
