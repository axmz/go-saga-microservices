package service

import (
	"context"
	"log/slog"

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
	slog.Info("InventoryReservationRequested:", "orderID", event.Id)
	if err := s.Repo.ReserveItems(ctx, event); err != nil {
		s.Kafka.PublishInventoryReservationFailedEvent(event.Id)
		return err
	}
	s.Kafka.PublishInventoryReservationSucceededEvent(event.Id)
	return nil
}

func (s *Service) ReleaseItems(ctx context.Context) {
	// return s.Repo.ReleaseItems(orderID, productID)
}

func (s *Service) MarkItemsSold(ctx context.Context, orderID string) {
	slog.Warn("Payment Succeeded:", "orderID", orderID)
	if err := s.Repo.MarkItemsSold(ctx, orderID); err != nil {

	}
}

func (s *Service) ReleaseReservedItems(ctx context.Context, orderID string) {
	slog.Warn("Payment Failed:", "orderID", orderID)
	if err := s.Repo.ReleaseReservedItems(ctx, orderID); err != nil {

	}
}
