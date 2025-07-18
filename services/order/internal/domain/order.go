package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	StatusPending         = "Pending"
	StatusAwaitingPayment = "AwaitingPayment"
	StatusPaid            = "Paid"
	StatusFailed          = "Failed"
)

type Order struct {
	ID        string      `json:"id"`
	Items     []OrderItem `json:"items"`
	Status    string      `json:"status"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

type OrderItem struct {
	ProductID string `json:"product_id"`
}

func NewOrder(items []OrderItem) *Order {
	id := uuid.New().String()
	now := time.Now()
	return &Order{
		ID:        id,
		Items:     items,
		Status:    StatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
