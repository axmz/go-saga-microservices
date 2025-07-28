package service

import (
	"context"

	"github.com/axmz/go-saga-microservices/inventory-service/internal/domain"
	"github.com/axmz/go-saga-microservices/inventory-service/internal/publisher"
	"github.com/axmz/go-saga-microservices/inventory-service/internal/repository"
	"github.com/axmz/go-saga-microservices/pkg/events"
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

func (s *Service) GetProducts(ctx context.Context) ([]domain.Product, error) {
	return s.Repo.GetProducts(ctx)
}

func (s *Service) ReserveItems(ctx context.Context, event *events.OrderCreatedEvent) error {
	if err := s.Repo.ReserveItems(ctx, event.Items); err != nil {
		s.Kafka.PublishReserveProductsEvent(event.Id, events.EventStatus_STATUS_FAILED)
		return err
	}
	s.Kafka.PublishReserveProductsEvent(event.Id, events.EventStatus_STATUS_SUCCESS)
	return nil
}

func (s *Service) ReleaseItems(ctx context.Context) {
	// return s.Repo.ReleaseItems(orderID, productID)
}
