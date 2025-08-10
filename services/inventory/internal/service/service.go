package service

import (
	"context"
	"log/slog"

	"github.com/axmz/go-saga-microservices/inventory-service/internal/domain"
	"github.com/axmz/go-saga-microservices/inventory-service/internal/publisher"
	"github.com/axmz/go-saga-microservices/inventory-service/internal/repository"
	"github.com/axmz/go-saga-microservices/pkg/proto/events"
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

func (s *Service) ReserveItems(ctx context.Context, event *events.OrderCreatedEvent) {
	if err := s.Repo.ReserveItems(ctx, event); err != nil {
		s.Kafka.PublishInventoryReservationFailedEvent(event.Id)
	}
	s.Kafka.PublishInventoryReservationSucceededEvent(event.Id)
}

func (s *Service) MarkItemsSold(ctx context.Context, orderID string) {
	if err := s.Repo.MarkItemsSold(ctx, orderID); err != nil {
		slog.Error("Failed to mark items as sold", "orderID", orderID, "err", err)
		return
	}
}

func (s *Service) ReleaseReservedItems(ctx context.Context, orderID string) {
	if err := s.Repo.ReleaseReservedItems(ctx, orderID); err != nil {
		slog.Error("Failed to release reserved items", "orderID", orderID, "err", err)
		return
	}
}
