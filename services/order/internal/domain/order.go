package domain

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusPending         Status = "Pending"
	StatusAwaitingPayment Status = "AwaitingPayment"
	StatusPaid            Status = "Paid"
	StatusFailed          Status = "Failed"
)

var ErrOrderNotFound = errors.New("order not found")

type ErrOrderNotFoundWithID struct {
	OrderID string
}

func (e *ErrOrderNotFoundWithID) Error() string {
	return fmt.Sprintf("order not found: %s", e.OrderID)
}

func (e *ErrOrderNotFoundWithID) Unwrap() error {
	return ErrOrderNotFound
}

func NewErrOrderNotFound(id string) error {
	return &ErrOrderNotFoundWithID{OrderID: id}
}

type Order struct {
	ID        string    `json:"id"`
	Items     []Item    `json:"items"`
	Status    Status    `json:"status"`
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
