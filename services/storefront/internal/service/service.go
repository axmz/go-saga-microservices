package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/axmz/go-saga-microservices/config"
	"github.com/axmz/go-saga-microservices/services/storefront/internal/domain"
)

type Service struct {
	cfg *config.Config
}

func New(cfg *config.Config) *Service {
	return &Service{
		cfg: cfg,
	}
}

func (s *Service) GetOrder(ctx context.Context, orderID string) (*domain.Order, error) {
	OrderServiceURL := s.cfg.Order.HTTP.URL()
	resp, err := http.Get(OrderServiceURL + "/orders?orderId=" + orderID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("order service returned status: %d", resp.StatusCode)
	}
	var order domain.Order
	if err := json.NewDecoder(resp.Body).Decode(&order); err != nil {
		return nil, err
	}
	return &order, nil
}

func (s *Service) GetAllProducts(ctx context.Context) ([]domain.Product, error) {
	// Call inventory service to get available products
	InventoryServiceURL := s.cfg.Inventory.HTTP.URL()
	resp, err := http.Get(InventoryServiceURL + "/products")
	if err != nil {
		return nil, fmt.Errorf("failed to call inventory service: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("inventory service returned status: %d", resp.StatusCode)
	}

	var products []domain.Product
	if err := json.NewDecoder(resp.Body).Decode(&products); err != nil {
		return nil, fmt.Errorf("failed to decode products: %v", err)
	}

	return products, nil
}

func (s *Service) CreateOrder(ctx context.Context, orderReq domain.CreateOrderRequest) error {
	jsonData, err := json.Marshal(orderReq)
	if err != nil {
		return err
	}

	// Send to order service
	OrderServiceURL := s.cfg.Order.HTTP.URL()
	resp, err := http.Post(OrderServiceURL+"/orders", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("order service returned status: %d", resp.StatusCode)
	}

	return nil
}
