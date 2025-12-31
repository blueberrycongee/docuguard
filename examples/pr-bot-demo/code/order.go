package demo

import (
	"errors"
	"time"
)

// OrderStatus represents the status of an order.
type OrderStatus string

const (
	StatusPending   OrderStatus = "pending"
	StatusConfirmed OrderStatus = "confirmed"
	StatusShipped   OrderStatus = "shipped"
	StatusDelivered OrderStatus = "delivered"
	StatusCancelled OrderStatus = "cancelled"
)

// Order represents a customer order.
type Order struct {
	ID         int
	UserID     int
	Items      []OrderItem
	TotalPrice float64
	Status     OrderStatus
	CreatedAt  time.Time
}

// OrderItem represents an item in an order.
type OrderItem struct {
	ProductID int
	Quantity  int
	Price     float64
}

// CreateOrder creates a new order for a user.
func CreateOrder(userID int, items []OrderItem) (*Order, error) {
	if userID <= 0 {
		return nil, errors.New("invalid user ID")
	}
	if len(items) == 0 {
		return nil, errors.New("order must contain at least one item")
	}

	total := calculateTotal(items)
	order := &Order{
		ID:         generateOrderID(),
		UserID:     userID,
		Items:      items,
		TotalPrice: total,
		Status:     StatusPending,
		CreatedAt:  time.Now(),
	}
	return order, nil
}

// GetOrder retrieves an order by its ID.
// Returns the order and an error if not found.
func GetOrder(orderID int) (*Order, error) {
	if orderID <= 0 {
		return nil, errors.New("invalid order ID")
	}
	// Simulate database lookup
	return &Order{
		ID:         orderID,
		UserID:     1,
		TotalPrice: 99.99,
		Status:     StatusConfirmed,
		CreatedAt:  time.Now(),
	}, nil
}

// UpdateOrderStatus updates the status of an order.
func UpdateOrderStatus(orderID int, status OrderStatus) error {
	if orderID <= 0 {
		return errors.New("invalid order ID")
	}
	// Validate status transition
	// Simulate update operation
	return nil
}

// CancelOrder cancels an order if it hasn't been shipped yet.
func CancelOrder(orderID int) error {
	order, err := GetOrder(orderID)
	if err != nil {
		return err
	}
	if order.Status == StatusShipped || order.Status == StatusDelivered {
		return errors.New("cannot cancel shipped or delivered order")
	}
	return UpdateOrderStatus(orderID, StatusCancelled)
}

func calculateTotal(items []OrderItem) float64 {
	total := 0.0
	for _, item := range items {
		total += item.Price * float64(item.Quantity)
	}
	return total
}

func generateOrderID() int {
	// Simulate ID generation
	return 67890
}
