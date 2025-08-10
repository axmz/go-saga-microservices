package service

import (
	"context"

	"github.com/axmz/go-saga-microservices/config"
	httppb "github.com/axmz/go-saga-microservices/pkg/proto/http"
	"github.com/axmz/go-saga-microservices/services/storefront/internal/client"
)

type Service struct {
	cfg             *config.Config
	orderClient     client.OrderClient
	paymentClient   client.PaymentClient
	inventoryClient client.InventoryClient
}

func New(cfg *config.Config, orderClient client.OrderClient, paymentClient client.PaymentClient, inventoryClient client.InventoryClient) *Service {
	return &Service{
		cfg:             cfg,
		orderClient:     orderClient,
		paymentClient:   paymentClient,
		inventoryClient: inventoryClient,
	}
}

func (s *Service) GetOrder(ctx context.Context, orderID string) (*httppb.Order, error) {
	resp, err := s.orderClient.GetOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}
	return resp.Order, nil
}

func (s *Service) GetProducts(ctx context.Context) ([]*httppb.Product, error) {
	resp, err := s.inventoryClient.GetProducts(ctx)
	if err != nil {
		return nil, err
	}
	return resp.Products, nil
}

func (s *Service) CreateOrder(ctx context.Context, orderReq *httppb.CreateOrderRequest) (*httppb.Order, error) {
	resp, err := s.orderClient.CreateOrder(ctx, orderReq)
	if err != nil {
		return nil, err
	}

	return resp.Order, nil
}

func (s *Service) PaymentSuccess(ctx context.Context, orderID string) error {
	protoReq := &httppb.PaymentSuccessRequest{OrderId: orderID}
	return s.paymentClient.PaymentSuccess(ctx, protoReq)
}

func (s *Service) PaymentFail(ctx context.Context, orderID string) error {
	protoReq := &httppb.PaymentFailRequest{OrderId: orderID}
	return s.paymentClient.PaymentFail(ctx, protoReq)
}
