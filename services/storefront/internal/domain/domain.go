package domain

import "time"

type Product struct {
	ID     int     `json:"id"`
	Name   string  `json:"name"`
	SKU    string  `json:"sku"`
	Status string  `json:"status"`
	Price  float64 `json:"price"`
}

type OrderItem struct {
	ProductID string `json:"product_id"`
}

type Order struct {
	ID        string      `json:"id"`
	Status    string      `json:"status"`
	Items     []OrderItem `json:"items"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

type PaymentInfo struct {
	CardNumber string `json:"card_number"`
	ExpiryDate string `json:"expiry_date"`
	CVV        string `json:"cvv"`
	Name       string `json:"name"`
}

type CreateOrderRequest struct {
	Items []OrderItem `json:"items"`
}
