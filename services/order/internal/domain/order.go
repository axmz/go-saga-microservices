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
	ID        string    `json:"id"`
	Items     []Item    `json:"items"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Item struct {
	ProductID string `json:"product_id"`
}

func NewOrder(items []Item) *Order {
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
