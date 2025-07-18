package service

import (
	"context"

	"github.com/axmz/go-saga-microservices/services/order/internal/domain"
	"github.com/axmz/go-saga-microservices/services/order/internal/publisher"
	"github.com/axmz/go-saga-microservices/services/order/internal/repository"
)

type Service struct {
	Repo  *repository.Repository
	Kafka *publisher.Publisher
}

func New(repo *repository.Repository, kafka *publisher.Publisher) *Service {
	return &Service{
		Repo:  repo,
		Kafka: kafka,
	}
}

func (s *Service) CreateOrder(ctx context.Context, order *domain.Order) error {
	return s.Repo.CreateOrder(ctx, order)
}

func (s *Service) GetOrder(ctx context.Context, orderId string) (*domain.Order, error) {
	return s.Repo.GetOrder(ctx, orderId)
}
