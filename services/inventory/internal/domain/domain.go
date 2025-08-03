package domain

type Status string

const (
	StatusAvailable Status = "available"
	StatusReserved  Status = "reserved"
	StatusSold      Status = "sold"
)

type Product struct {
	ID     int     `json:"id"`
	Name   string  `json:"name"`
	SKU    string  `json:"sku"`
	Status string  `json:"status"`
	Price  float64 `json:"price"`
}

func NewProduct(name, sku, status string, price float64) *Product {
	return &Product{
		Name:   name,
		SKU:    sku,
		Status: status,
		Price:  price,
	}
}
